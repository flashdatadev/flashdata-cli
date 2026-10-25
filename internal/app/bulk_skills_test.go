package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/flashdatadev/flashdata-cli/internal/agentskills"
	"github.com/flashdatadev/flashdata-cli/internal/state"
)

// Execute the shipped guides, including their option assets, against mock APIs.
// This catches invalid flags/options and recovery examples that repeat paid work.
func TestBundledBulkSkillWorkflows(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	installed, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := agentskills.Install([]string{installed}, nil); err != nil {
		t.Fatal(err)
	}
	storage, err := os.ReadFile(filepath.Join("..", "..", "examples", "storage.example.json"))
	if err != nil {
		t.Fatal(err)
	}
	blocks := regexp.MustCompile("(?s)```([^\\n]*)\\n(.*?)```")
	for _, skill := range []string{"flashdata", "youtube-metadata", "youtube-transcript", "youtube-captions", "youtube-license-check", "youtube-download"} {
		guide, err := os.ReadFile(filepath.Join(installed, skill, "references", "bulk.md"))
		if err != nil {
			t.Fatal(err)
		}
		var commands []string
		var ids, inlineOptions []byte
		var recovery string
		for _, block := range blocks.FindAllStringSubmatch(string(guide), -1) {
			switch block[1] {
			case "text":
				ids = []byte(block[2])
			case "json":
				inlineOptions = []byte(block[2])
			case "sh":
				for _, line := range strings.Split(block[2], "\n") {
					if strings.HasPrefix(line, "flashdata bulk run ") {
						commands = append(commands, line)
					} else if strings.HasPrefix(line, "flashdata runs recover ") {
						recovery = line
					}
				}
			}
		}
		if len(commands) == 0 || len(ids) == 0 || recovery == "" {
			t.Fatal("bulk guide needs executable input, submission and recovery examples", skill)
		}
		for _, command := range commands {
			for _, outcome := range []string{"success", "failed-item", "partial-transfer"} {
				t.Run(skill+"/"+command+"/"+outcome, func(t *testing.T) {
					base, err := filepath.EvalSymlinks(t.TempDir())
					if err != nil {
						t.Fatal(err)
					}
					work := filepath.Join(base, "work")
					write := func(name string, data []byte) {
						t.Helper()
						if err := state.AtomicWrite(filepath.Join(work, name), data); err != nil {
							t.Fatal(err)
						}
					}
					write("ids.txt", ids)
					write("storage.json", storage)
					if len(inlineOptions) > 0 {
						write("audio-options.json", inlineOptions)
					}
					for _, names := range [][2]string{{"bulk-options.json", "options.json"}, {"bulk-audio-options.json", "audio-options.json"}, {"bulk-video-options.json", "video-options.json"}} {
						data, err := os.ReadFile(filepath.Join(installed, skill, "assets", names[0]))
						if os.IsNotExist(err) {
							continue
						}
						if err != nil {
							t.Fatal(err)
						}
						write(names[1], data)
					}
					parse := func(line string) []string {
						t.Helper()
						// These examples use literal arguments, never a shell.
						if strings.ContainsAny(line, "|><;$`\\\"") {
							t.Fatal("bulk guide needs literal arguments", line)
						}
						args := strings.Fields(strings.TrimPrefix(line, "flashdata "))
						for i, arg := range args {
							if strings.HasPrefix(arg, "/absolute/work/") {
								args[i] = filepath.Join(work, strings.TrimPrefix(arg, "/absolute/work/"))
							}
						}
						return args
					}
					args := parse(command)
					flag := func(name string) string {
						t.Helper()
						for i, arg := range args {
							if arg == name && i+1 < len(args) {
								return args[i+1]
							}
						}
						t.Fatal("missing example flag", name)
						return ""
					}
					m, server := fixture(t)
					root := filepath.Join(base, "state")
					login(t, root, server.URL, "fd_owner_offline")
					code, out, stderr := invoke(t, root, server.URL, "", append(append([]string{}, args...), "--dry-run")...)
					if code != 0 || m.stats().CreateCount != 0 || m.stats().UploadCount != 0 || m.stats().StartCount != 0 {
						t.Fatal("dry run submitted work or rejected documented inputs", out, stderr, m.stats())
					}
					m.set(func(m *mockAPI) {
						m.Partial = outcome == "failed-item"
						m.TruncatePart = outcome == "partial-transfer"
					})
					code, out, stderr = invoke(t, root, server.URL, "", args...)
					if (code == 0) != (outcome == "success") {
						t.Fatal("wrong status for bulk outcome", out, stderr)
					}
					exported, err := os.ReadFile(flag("--output"))
					if err != nil {
						t.Fatal(err, stderr)
					}
					checkRows := func(data []byte) {
						t.Helper()
						rows := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
						for _, row := range rows {
							var value map[string]json.RawMessage
							if json.Unmarshal(row, &value) != nil || value["index"] == nil || value["status"] == nil {
								t.Fatal("export is not per-item JSONL", string(data))
							}
						}
						if !bytes.Contains(data, []byte("9007199254740993123")) {
							t.Fatal("JSONL changed a large source integer", string(data))
						}
					}
					checkRows(exported)
					if outcome == "failed-item" && !bytes.Contains(exported, []byte(`"status":"invalid"`)) {
						t.Fatal("failed item was omitted", string(exported))
					}
					runID := flag("--run-id")
					run, err := stateStore(t, root).ReadRun(runID)
					if err != nil || run == nil || run.RemoteID != "bulk_offline1" {
						t.Fatal("missing recoverable job ID", run, err)
					}
					// The guide tells users to substitute the run ID for each variant.
					recoverArgs := parse(recovery)
					recoverArgs[2] = runID
					m.set(func(m *mockAPI) { m.TruncatePart = false })
					code, out, stderr = invoke(t, root, server.URL, "", recoverArgs...)
					if (code == 0) != (outcome != "failed-item") {
						t.Fatal("recovery hid failures or failed to complete a transfer", out, stderr)
					}
					recovered, err := os.ReadFile(filepath.Join(work, "recovered.jsonl"))
					if err != nil {
						t.Fatal(err, stderr)
					}
					checkRows(recovered)
					if outcome == "failed-item" && !bytes.Contains(recovered, []byte(`"status":"invalid"`)) {
						t.Fatal("recovery hid a failed item")
					}
					stats := m.stats()
					if stats.CreateCount != 1 || stats.UploadCount != 1 || stats.StartCount != 1 || stats.QueryCount != 0 {
						t.Fatal("recovery repeated work", stats)
					}
				})
			}
		}
	}
}
