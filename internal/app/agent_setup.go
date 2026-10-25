package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flashdatadev/flashdata-cli/internal/agentskills"
	"github.com/spf13/cobra"
)

type skillFlags struct {
	Agents, Skills []string
	Global         bool
	Directory      string
}

func bindSkills(cmd *cobra.Command, flags *skillFlags) {
	cmd.Flags().StringSliceVar(&flags.Agents, "agent", nil, "codex, claude-code or cursor; comma-separated")
	cmd.Flags().StringSliceVar(&flags.Skills, "skill", nil, "selected skill names; default: all 16 CLI skills")
	cmd.Flags().BoolVar(&flags.Global, "global", false, "install for the current user instead of this project")
	cmd.Flags().StringVar(&flags.Directory, "directory", "", "project directory; defaults to the working directory")
	cmd.MarkFlagsMutuallyExclusive("global", "directory")
}

func (a *App) addAgentSetup() {
	group := &cobra.Command{Use: "skills", Short: "List, install and verify bundled agent skills offline"}
	a.Root.AddCommand(group)
	list := &cobra.Command{Use: "list", Short: "List bundled skills and their source revision", Args: cobra.NoArgs}
	list.RunE = func(cmd *cobra.Command, args []string) error {
		return a.success(map[string]any{"skills": agentskills.Names(), "sourceCommit": agentskills.Source(), "agents": []string{"codex", "claude-code", "cursor"}, "mcpOptional": true}, "", false, nil)
	}
	group.AddCommand(list)
	for _, action := range []string{"install", "verify"} {
		flags := &skillFlags{}
		command := &cobra.Command{Use: action, Short: action + " bundled skills for selected agents", Args: cobra.NoArgs}
		bindSkills(command, flags)
		command.RunE = func(cmd *cobra.Command, args []string) error {
			roots, names, err := a.skillTargets(cmd, *flags)
			if err != nil {
				return err
			}
			if action == "verify" {
				if err = agentskills.Verify(roots, names); err != nil {
					return skillError(err)
				}
				return a.success(map[string]any{"verified": true, "directories": roots, "skills": names, "sourceCommit": agentskills.Source()}, "", false, nil)
			}
			results, err := agentskills.Install(roots, names)
			if err != nil {
				return skillError(err)
			}
			return a.success(map[string]any{"verified": true, "installations": results, "sourceCommit": agentskills.Source()}, "", false, nil)
		}
		group.AddCommand(command)
	}
	flags := &skillFlags{}
	var name, audience, client, scopes string
	var skipAuth bool
	setup := &cobra.Command{Use: "setup", Short: "Select agents, install skills and verify FlashData access", Args: cobra.NoArgs}
	bindSkills(setup, flags)
	setup.Flags().StringVar(&name, "name", "FlashData agent", "name for new temporary access")
	setup.Flags().StringVar(&audience, "audience", "coding", "coding or personal")
	setup.Flags().StringVar(&client, "client", "flashdata-cli", "self-reported client label")
	setup.Flags().StringVar(&scopes, "scopes", "google_search,youtube_metadata", "comma-separated trial source names")
	setup.Flags().BoolVar(&skipAuth, "skip-auth", false, "install and verify skills without contacting FlashData")
	setup.RunE = func(cmd *cobra.Command, args []string) error {
		roots, names, err := a.skillTargets(cmd, *flags)
		if err != nil {
			return err
		}
		// Resolve conflicts before creating temporary access or changing credentials.
		if err = agentskills.Check(roots, names); err != nil {
			return skillError(err)
		}
		var identity any
		if !skipAuth {
			if _, err = a.initializeAccess(cmd, name, audience, client, scopes); err != nil {
				return err
			}
			ctx, cancel := a.context(cmd)
			defer cancel()
			store, err := a.store()
			if err != nil {
				return err
			}
			vault, err := a.load(ctx, store)
			if err != nil {
				return err
			}
			access, err := a.authorize(ctx, vault, "", false)
			if err != nil {
				return err
			}
			identity = publicIdentity(access)
		}
		results, err := agentskills.Install(roots, names)
		if err != nil {
			return skillError(err)
		}
		return a.success(map[string]any{
			"skillsVerified": true, "accessVerified": !skipAuth, "identity": identity,
			"installations": results, "sourceCommit": agentskills.Source(), "paidQueries": 0,
			"nextCommand": "flashdata sources list",
		}, "", false, nil)
	}
	a.Root.AddCommand(setup)
}

func skillError(err error) error {
	return fail("skill_installation_failed", err.Error(), "Resolve the reported file or directory, then rerun the same command.")
}

func (a *App) skillTargets(cmd *cobra.Command, flags skillFlags) ([]string, []string, error) {
	names, err := agentskills.Select(flags.Skills)
	if err != nil {
		return nil, nil, fail("invalid_skill", err.Error(), "")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}
	directory := flags.Directory
	if directory == "" {
		directory, err = os.Getwd()
	}
	if err != nil {
		return nil, nil, err
	}
	directory, err = filepath.Abs(directory)
	if err != nil {
		return nil, nil, err
	}
	if len(flags.Agents) == 0 {
		if !a.prompting() {
			return nil, nil, fail("agent_required", "Choose an agent for skill installation.", "Use --agent codex, --agent claude-code or --agent cursor.")
		}
		detected := detectSkillAgents(home, directory)
		_, _ = fmt.Fprintln(a.IO.Err, "Choose agents: codex, claude-code, cursor (comma-separated).")
		if len(detected) > 0 {
			_, _ = fmt.Fprintf(a.IO.Err, "Detected: %s. Press Enter to select these.\n", strings.Join(detected, ", "))
		}
		_, _ = fmt.Fprint(a.IO.Err, "Agents: ")
		line, err := a.promptLine(cmd)
		if err != nil {
			return nil, nil, err
		}
		if strings.TrimSpace(line) == "" {
			flags.Agents = detected
		} else {
			flags.Agents = strings.Split(strings.TrimSpace(line), ",")
		}
		if len(flags.Agents) == 0 {
			return nil, nil, fail("agent_required", "No agent was selected.", "Rerun with --agent NAME.")
		}
	}
	roots, err := agentSkillRoots(home, directory, flags.Agents, flags.Global)
	if err != nil {
		return nil, nil, fail("invalid_agent", err.Error(), "")
	}
	if a.prompting() && cmd.Name() != "verify" {
		_, _ = fmt.Fprintf(a.IO.Err, "Installing %d skills in:\n%s\n", len(names), strings.Join(roots, "\n"))
		_, _ = fmt.Fprint(a.IO.Err, "Continue? [y/N]: ")
		answer, err := a.promptLine(cmd)
		if err != nil {
			return nil, nil, err
		}
		if !strings.EqualFold(strings.TrimSpace(answer), "y") && !strings.EqualFold(strings.TrimSpace(answer), "yes") {
			return nil, nil, fail("setup_cancelled", "No skills were changed.", "")
		}
	}
	return roots, names, nil
}

// Read one line without buffering the next prompt's input. The command deadline
// also bounds an unattended or disconnected terminal prompt.
func (a *App) promptLine(cmd *cobra.Command) (string, error) {
	type result struct {
		text string
		err  error
	}
	ready := make(chan result, 1)
	go func() {
		var b strings.Builder
		one := make([]byte, 1)
		for b.Len() < 4096 {
			n, err := a.IO.In.Read(one)
			if n > 0 {
				if one[0] == '\n' {
					ready <- result{b.String(), nil}
					return
				}
				b.WriteByte(one[0])
			}
			if err != nil {
				ready <- result{b.String(), err}
				return
			}
		}
		ready <- result{"", fmt.Errorf("input is too long")}
	}()
	select {
	case <-cmd.Context().Done():
		return "", cmd.Context().Err()
	case value := <-ready:
		return value.text, value.err
	}
}

func detectSkillAgents(home, project string) []string {
	var found []string
	paths := map[string][]string{"codex": {filepath.Join(home, ".codex"), filepath.Join(project, ".codex")}, "claude-code": {filepath.Join(home, ".claude"), filepath.Join(project, ".claude")}, "cursor": {filepath.Join(home, ".cursor"), filepath.Join(project, ".cursor")}}
	if value := os.Getenv("CODEX_HOME"); value != "" {
		paths["codex"] = append(paths["codex"], value)
	}
	if value := os.Getenv("CLAUDE_CONFIG_DIR"); value != "" {
		paths["claude-code"] = append(paths["claude-code"], value)
	}
	for _, name := range []string{"codex", "claude-code", "cursor"} {
		for _, path := range paths[name] {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				found = append(found, name)
				break
			}
		}
	}
	return found
}

func agentSkillRoots(home, project string, agents []string, global bool) ([]string, error) {
	// Resolve the selected base (including macOS /var) before appending agent
	// directories. Links within the skill destination itself are still rejected.
	var err error
	home, err = canonicalSkillBase(home)
	if err != nil {
		return nil, err
	}
	project, err = canonicalSkillBase(project)
	if err != nil {
		return nil, err
	}
	unique := map[string]bool{}
	for _, agent := range agents {
		var root string
		switch strings.TrimSpace(agent) {
		case "codex":
			root = filepath.Join(project, ".agents", "skills")
			if global {
				root = filepath.Join(home, ".agents", "skills")
			}
		case "cursor":
			root = filepath.Join(project, ".agents", "skills")
			if global {
				root = filepath.Join(home, ".cursor", "skills")
			}
		case "claude-code":
			root = filepath.Join(project, ".claude", "skills")
			if global {
				config := os.Getenv("CLAUDE_CONFIG_DIR")
				if config == "" {
					config = filepath.Join(home, ".claude")
				}
				config, err = canonicalSkillBase(config)
				if err != nil {
					return nil, err
				}
				root = filepath.Join(config, "skills")
			}
		default:
			return nil, fmt.Errorf("unsupported agent %q; choose codex, claude-code or cursor", agent)
		}
		abs, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		unique[abs] = true
	}
	roots := []string{}
	for root := range unique {
		roots = append(roots, root)
	}
	sort.Strings(roots)
	return roots, nil
}

func canonicalSkillBase(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	ancestor := abs
	for {
		if _, err = os.Lstat(ancestor); err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", err
		}
		ancestor = parent
	}
	resolved, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(ancestor, abs)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolved, rel), nil
}
