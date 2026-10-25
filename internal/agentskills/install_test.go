package agentskills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallVerifyUpdateAndProtectEdits(t *testing.T) {
	root := realTemp(t)
	results, err := Install([]string{root}, nil)
	if err != nil || len(results) != 16 {
		t.Fatal(results, err)
	}
	if _, err = os.Stat(filepath.Join(root, "flashdata-mcp")); !os.IsNotExist(err) {
		t.Fatal("optional MCP installed by default")
	}
	if err = Verify([]string{root}, nil); err != nil {
		t.Fatal(err)
	}
	results, err = Install([]string{root}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.Status != "unchanged" {
			t.Fatal(r)
		}
	}
	path := filepath.Join(root, "google-search", "SKILL.md")
	if err = os.WriteFile(path, []byte("my edits"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err = Install([]string{root}, nil); err == nil {
		t.Fatal("overwrote local edit")
	}
	actual, _ := os.ReadFile(path)
	if string(actual) != "my edits" {
		t.Fatal("local edit lost")
	}
	if Verify([]string{root}, nil) == nil {
		t.Fatal("verification accepted changed skill")
	}
}

func TestManagedUpgradeAndUntrackedFiles(t *testing.T) {
	root := realTemp(t)
	names := []string{"google-search"}
	if _, err := Install([]string{root}, names); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, names[0])
	record, _, _ := readReceipt(dir)
	old := []byte("previous released skill")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), old, 0644); err != nil {
		t.Fatal(err)
	}
	record.Files["SKILL.md"] = digest(old)
	record.Files["obsolete.md"] = digest(old)
	_ = os.WriteFile(filepath.Join(dir, "obsolete.md"), old, 0644)
	_ = os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("keep my notes"), 0644)
	b, _ := json.Marshal(record)
	_ = os.WriteFile(filepath.Join(dir, receiptName), b, 0644)
	results, err := Install([]string{root}, names)
	if err != nil || results[0].Status != "updated" {
		t.Fatal(results, err)
	}
	if _, err = os.Stat(filepath.Join(dir, "obsolete.md")); !os.IsNotExist(err) {
		t.Fatal("stale managed file kept")
	}
	notes, _ := os.ReadFile(filepath.Join(dir, "notes.txt"))
	if string(notes) != "keep my notes" {
		t.Fatal("untracked file changed")
	}
	if err = Verify([]string{root}, names); err != nil {
		t.Fatal(err)
	}
}

func TestConflictsPreflightAllDestinations(t *testing.T) {
	first, second := realTemp(t), realTemp(t)
	_ = os.MkdirAll(filepath.Join(second, "google-search"), 0755)
	_ = os.WriteFile(filepath.Join(second, "google-search", "SKILL.md"), []byte("another provider"), 0644)
	if _, err := Install([]string{first, second}, []string{"google-search"}); err == nil {
		t.Fatal("conflicting skill accepted")
	}
	entries, _ := os.ReadDir(first)
	if len(entries) != 0 {
		t.Fatal("first destination changed before conflict detected")
	}
	if _, err := Install([]string{first}, []string{"../../escape"}); err == nil {
		t.Fatal("unsafe skill selected")
	}
}

func TestMaliciousReceiptCannotEscapeSkill(t *testing.T) {
	root := realTemp(t)
	dir := filepath.Join(root, "google-search")
	_ = os.MkdirAll(dir, 0755)
	outside := filepath.Join(root, "keep.txt")
	_ = os.WriteFile(outside, []byte("private"), 0600)
	for _, name := range []string{"../keep.txt", "C:/keep.txt", "..\\keep.txt", receiptName} {
		b, _ := json.Marshal(Receipt{Version: 1, Files: map[string]string{name: digest([]byte("private"))}})
		_ = os.WriteFile(filepath.Join(dir, receiptName), b, 0600)
		if _, err := Install([]string{root}, []string{"google-search"}); err == nil {
			t.Fatal("unsafe record accepted", name)
		}
	}
	b, _ := os.ReadFile(outside)
	if string(b) != "private" {
		t.Fatal("escaped target")
	}
}

func TestLinkedDestinationAndConcurrentInstallFailClosed(t *testing.T) {
	root := realTemp(t)
	_ = os.WriteFile(filepath.Join(root, ".flashdata-skills.lock"), []byte("busy"), 0600)
	if _, err := Install([]string{root}, []string{"google-search"}); err == nil {
		t.Fatal("concurrent lock ignored")
	}
	if _, err := os.Stat(filepath.Join(root, ".flashdata-skills.lock")); err != nil {
		t.Fatal("another install lock removed")
	}
	link := filepath.Join(realTemp(t), "link")
	if err := os.Symlink(realTemp(t), link); err != nil {
		t.Skip("symlinks unavailable", err)
	}
	if _, err := Install([]string{link}, []string{"google-search"}); err == nil {
		t.Fatal("symbolic-link destination followed")
	}
}

func TestOptionalMCPAndPortableAssets(t *testing.T) {
	root := realTemp(t)
	if _, err := Install([]string{root}, []string{"flashdata-mcp", "youtube-transcript"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"flashdata-mcp/SKILL.md", "youtube-transcript/assets/query.json", "youtube-transcript/references/mcp.md", "youtube-transcript/scripts/mcp_response.py"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			t.Fatal(name, err)
		}
	}
}

func realTemp(t *testing.T) string {
	t.Helper()
	path, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return path
}
