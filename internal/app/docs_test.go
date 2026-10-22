package app

import (
	"encoding/json"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/flashdatadev/flashdata-cli/internal/state"
)

func TestDocumentationLanguagesLinksAndExamples(t *testing.T) {
	root := filepath.Join("..", "..")
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	en, zh := read("README.md"), read("README.zh-CN.md")
	if !strings.Contains(strings.SplitN(en, "\n", 5)[2], "简体中文") {
		t.Fatal("English default needs top Chinese link")
	}
	blocks := regexp.MustCompile("(?s)```[^\n]*\n.*?```")
	for _, pair := range [][2]string{{en, zh}, {read("docs/installation.md"), read("docs/installation.zh-CN.md")}, {read("docs/python-packaging.md"), read("docs/python-packaging.zh-CN.md")}, {read("packaging/pypi/README.md"), read("packaging/pypi/README.zh-CN.md")}} {
		a, b := blocks.FindAllString(pair[0], -1), blocks.FindAllString(pair[1], -1)
		if len(a) != len(b) {
			t.Fatal("translated command block count differs")
		}
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("translated command block %d drifted", i)
			}
		}
	}
	links := regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	inline := regexp.MustCompile("`[^`]*`")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".venv") {
				return filepath.SkipDir
			}
			switch entry.Name() {
			case ".git", ".cache", "dist":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(raw)
		if !strings.Contains(text, "https://flashdata.dev") {
			t.Error("missing official link", path)
		}
		text = inline.ReplaceAllString(blocks.ReplaceAllString(text, ""), "")
		for _, m := range links.FindAllStringSubmatch(text, -1) {
			target := strings.Split(m[1], "#")[0]
			parsed, err := url.Parse(target)
			if err != nil {
				t.Error("invalid link", path, target)
				continue
			}
			if target == "" || parsed.Scheme != "" {
				continue
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), target)); err != nil {
				t.Error("broken link", path, target)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"query.json", "query-flat.json", "youtube-metadata.json"} {
		if _, _, err := normalizeQuery([]byte(read("examples/"+name)), false); err != nil {
			t.Fatal(name, err)
		}
	}
	for _, name := range []string{"batch.json", "bulk-options.json", "bulk-create.json", "bulk-update.json", "download-options.json", "storage.example.json"} {
		if !json.Valid([]byte(read("examples/" + name))) {
			t.Fatal("invalid example JSON", name)
		}
	}
}

// Execute the README's actual workflow commands and JSON inputs, rather than
// maintaining a second list of examples that could drift from the public docs.
func TestReadmeWorkflows(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	raw, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	blocks := regexp.MustCompile("(?s)```([^\\n]*)\\n(.*?)```").FindAllStringSubmatch(string(raw), -1)
	// macOS temporary directories can be reached through /var -> /private/var.
	// The private writer requires a canonical parent, just like state.New uses.
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(base, "work")
	for _, block := range blocks {
		var name string
		switch block[1] {
		case "json":
			var fields map[string]json.RawMessage
			if err := json.Unmarshal([]byte(block[2]), &fields); err != nil {
				t.Fatal("invalid README JSON", err)
			}
			switch {
			case fields["source"] != nil:
				name = "news-query.json"
			case fields["storage_type"] != nil:
				name = "storage.json"
			case fields["download_type"] != nil:
				name = "audio-options.json"
			default:
				t.Fatal("README JSON needs an associated runnable workflow")
			}
		case "text":
			name = "ids.txt"
		default:
			continue
		}
		if err := state.AtomicWrite(filepath.Join(dir, name), []byte(block[2])); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	mock, server := fixture(t)
	root := filepath.Join(dir, "state")
	// README examples use only literal words and double-quoted query strings;
	// reject shell syntax rather than ever executing a reader-provided shell.
	words := regexp.MustCompile(`"[^"]*"|[^\s"]+`)
	commands, outputs := 0, 0
	for _, block := range blocks {
		if block[1] != "bash" {
			continue
		}
		for _, line := range strings.Split(block[2], "\n") {
			if !strings.HasPrefix(line, "flashdata ") {
				continue
			}
			if strings.ContainsAny(line, "|><;$`\\") {
				t.Fatal("README workflow needs a literal command", line)
			}
			args := words.FindAllString(strings.TrimPrefix(line, "flashdata "), -1)
			for i := range args {
				args[i] = strings.Trim(args[i], `"`)
			}
			input := ""
			if line == "flashdata auth login" {
				// The test has no terminal. Import the same dummy key through
				// the documented stdin route instead of the hidden prompt.
				args = append(args, "--api-key-stdin")
				input = "fd_owner_offline\n"
			}
			code, _, stderr := invoke(t, root, server.URL, input, args...)
			if code != 0 {
				t.Fatalf("README command failed: %s\n%s", line, stderr)
			}
			commands++
			for i, arg := range args {
				if arg != "--output" || i+1 == len(args) {
					continue
				}
				output, err := os.ReadFile(args[i+1])
				if err != nil || len(output) == 0 {
					t.Fatal("README command did not save its output", line, err)
				}
				for _, row := range strings.Split(strings.TrimSpace(string(output)), "\n") {
					if strings.HasSuffix(args[i+1], ".jsonl") && !json.Valid([]byte(row)) {
						t.Fatal("invalid README JSONL export", line)
					}
				}
				if !strings.HasSuffix(args[i+1], ".jsonl") && !json.Valid(output) {
					t.Fatal("invalid README JSON export", line)
				}
				outputs++
			}
		}
	}
	stats := mock.stats()
	if commands == 0 || outputs == 0 || stats.CreateCount != 2 || stats.UploadCount != 2 || stats.StartCount != 2 {
		t.Fatalf("README did not exercise both complete bulk workflows and recovery: commands=%d outputs=%d stats=%+v", commands, outputs, stats)
	}
}
