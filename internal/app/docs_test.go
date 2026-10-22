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
