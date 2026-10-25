package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/flashdatadev/flashdata-cli/internal/agentskills"
)

func TestGuidedSetupReusesAccessAndDoesNotQuery(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root, project := filepath.Join(t.TempDir(), "state"), t.TempDir()
	code, out, errout := invoke(t, root, server.URL, "", "setup", "--agent", "codex,cursor", "--directory", project, "--name", "test agent")
	if code != 0 {
		t.Fatal(code, out, errout)
	}
	var result struct {
		Data struct {
			AccessVerified bool `json:"accessVerified"`
			SkillsVerified bool `json:"skillsVerified"`
			PaidQueries    int  `json:"paidQueries"`
		} `json:"data"`
	}
	if json.Unmarshal([]byte(out), &result) != nil || !result.Data.AccessVerified || !result.Data.SkillsVerified || result.Data.PaidQueries != 0 {
		t.Fatal(out)
	}
	if m.RegisterCount != 1 || m.QueryCount != 0 {
		t.Fatal(m.RegisterCount, m.QueryCount)
	}
	code, out, errout = invoke(t, root, server.URL, "", "setup", "--agent", "codex", "--directory", project)
	if code != 0 || m.RegisterCount != 1 || m.QueryCount != 0 {
		t.Fatal(out, errout, m.RegisterCount, m.QueryCount)
	}
	if _, err := os.Stat(filepath.Join(project, ".agents/skills/flashdata/SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(project, ".agents/skills/flashdata-mcp")); !os.IsNotExist(err) {
		t.Fatal("MCP installed")
	}
}

func TestOfflineSkillsAndSetupValidateBeforeAccess(t *testing.T) {
	for _, args := range [][]string{
		{"skills", "list"},
		{"skills", "install", "--agent", "claude-code"},
		{"setup", "--agent", "cursor", "--skip-auth"},
	} {
		project := t.TempDir()
		if args[1] != "list" {
			args = append(args, "--directory", project)
		}
		var out, stderr bytes.Buffer
		a := New(IO{In: strings.NewReader(""), Out: &out, Err: &stderr}, "test", "test")
		code := a.Execute(context.Background(), append(args, "--non-interactive"))
		if code != 0 {
			t.Fatal(out.String(), stderr.String())
		}
		if strings.Contains(out.String(), `"accessVerified":true`) {
			t.Fatal("offline installation claimed access verification")
		}
	}
	m, server := fixture(t)
	for _, args := range [][]string{{"setup"}, {"setup", "--agent", "invalid"}, {"setup", "--agent", "codex", "--skill", "invalid"}} {
		code, _, errout := invoke(t, t.TempDir(), server.URL, "", args...)
		if code == 0 || m.RegisterCount != 0 || m.AccessReads != 0 {
			t.Fatal(code, errout, m)
		}
	}
	project := t.TempDir()
	dir := filepath.Join(project, ".agents/skills/google-search")
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("local edits"), 0644)
	code, _, errout := invoke(t, t.TempDir(), server.URL, "", "setup", "--agent", "codex", "--directory", project)
	if code == 0 || m.RegisterCount != 0 || !strings.Contains(errout, "skill_installation_failed") {
		t.Fatal(code, errout)
	}
}

func TestInteractiveAgentChoiceAndCancellation(t *testing.T) {
	for _, answer := range []string{"codex,cursor\ny\n", "codex\nn\n"} {
		var out, stderr bytes.Buffer
		project := t.TempDir()
		a := New(IO{In: strings.NewReader(answer), Out: &out, Err: &stderr}, "test", "test")
		a.interactive = true
		code := a.Execute(context.Background(), []string{"setup", "--skip-auth", "--directory", project})
		if strings.Contains(answer, "\ny\n") {
			if code != 0 {
				t.Fatal(out.String(), stderr.String())
			}
		} else {
			if code == 0 {
				t.Fatal("cancel ignored")
			}
			entries, _ := os.ReadDir(project)
			if len(entries) != 0 {
				t.Fatal("cancelled setup wrote files")
			}
		}
	}
}

func TestAgentRootsAndDetection(t *testing.T) {
	home, project := t.TempDir(), t.TempDir()
	home, _ = filepath.EvalSymlinks(home)
	project, _ = filepath.EvalSymlinks(project)
	t.Setenv("CODEX_HOME", filepath.Join(home, "custom-codex"))
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, "custom-claude"))
	_ = os.MkdirAll(os.Getenv("CODEX_HOME"), 0755)
	_ = os.MkdirAll(os.Getenv("CLAUDE_CONFIG_DIR"), 0755)
	found := detectSkillAgents(home, project)
	if strings.Join(found, ",") != "codex,claude-code" {
		t.Fatal(found)
	}
	roots, err := agentSkillRoots(home, project, []string{"codex", "cursor"}, false)
	if err != nil || len(roots) != 1 || roots[0] != filepath.Join(project, ".agents/skills") {
		t.Fatal(roots, err)
	}
	roots, err = agentSkillRoots(home, project, []string{"claude-code"}, true)
	if err != nil || roots[0] != filepath.Join(home, "custom-claude/skills") {
		t.Fatal(roots, err)
	}
}

func TestBundledQueryAssetsAndPlainTextCommands(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	installed, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = agentskills.Install([]string{installed}, nil); err != nil {
		t.Fatal(err)
	}
	assets, err := filepath.Glob(filepath.Join(installed, "*", "assets", "query.json"))
	if err != nil || len(assets) != 14 {
		t.Fatal(assets, err)
	}
	for _, path := range assets {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err = normalizeQuery(b, false); err != nil {
			t.Fatal(path, err)
		}
	}
	_, server, wire := formatFixture(t)
	stateDir := filepath.Join(t.TempDir(), "state")
	login(t, stateDir, server.URL, "fd_owner_offline")
	words := regexp.MustCompile(`"[^"]*"|[^\s"]+`)
	for _, name := range []string{"google-search", "youtube-transcript"} {
		b, _ := os.ReadFile(filepath.Join(installed, name, "SKILL.md"))
		block := regexp.MustCompile("(?s)```sh\\n(.*?)\\n```").FindStringSubmatch(string(b))
		if len(block) != 2 {
			t.Fatal("missing runnable command", name)
		}
		args := words.FindAllString(strings.TrimPrefix(block[1], "flashdata "), -1)
		for i := range args {
			args[i] = strings.Trim(args[i], `"`)
		}
		code, out, stderr := invokeFormat(t, stateDir, server.URL, "", false, args...)
		if code != 0 || len(out) == 0 || json.Valid([]byte(out)) {
			t.Fatal(name, out, stderr)
		}
	}
	count := 0
	for _, call := range wire.calls() {
		if call.Path != "/v1/queries/realtime" {
			continue
		}
		count++
		if call.Accept != "text/plain" || call.RespondWith != "" {
			t.Fatal("skill did not request plain text", call)
		}
	}
	if count != 2 {
		t.Fatal("unexpected number of query submissions", count)
	}
}

func TestSetupRejectsRevokedAccessWithoutInstallingOrRegistering(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root, project := filepath.Join(t.TempDir(), "state"), t.TempDir()
	login(t, root, server.URL, "fd_owner_offline")
	m.set(func(m *mockAPI) { m.Revoked = true })
	code, _, stderr := invoke(t, root, server.URL, "", "setup", "--agent", "codex", "--directory", project)
	if code == 0 || !strings.Contains(stderr, "invalid_api_key") || m.stats().RegisterCount != 0 {
		t.Fatal(code, stderr)
	}
	entries, _ := os.ReadDir(project)
	if len(entries) != 0 {
		t.Fatal("installed despite failed access check")
	}
}

func TestStandaloneBinaryCarriesCompleteSkillBundle(t *testing.T) {
	if testing.Short() {
		t.Skip("binary acceptance")
	}
	binary := buildBinary(t)
	project, stateDir := t.TempDir(), filepath.Join(t.TempDir(), "unused-state")
	for _, args := range [][]string{
		{"setup", "--skip-auth", "--agent", "codex", "--directory", project},
		{"skills", "verify", "--agent", "codex", "--directory", project},
	} {
		code, out, stderr := binaryCall(t, binary, stateDir, "", "", args...)
		if code != 0 {
			t.Fatal(code, out, stderr)
		}
	}
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Fatal("offline setup touched credentials", err)
	}
}
