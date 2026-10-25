// Package agentskills installs the skill snapshot shipped with this executable.
// Installation is offline and never changes agent configuration or credentials.
package agentskills

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed bundle
var bundled embed.FS

const receiptName = ".flashdata-install.json"

type Receipt struct {
	Version int               `json:"version"`
	Source  string            `json:"source"`
	Files   map[string]string `json:"files"`
}

type Result struct {
	Skill     string `json:"skill"`
	Directory string `json:"directory"`
	Status    string `json:"status"`
}

func Names() []string {
	entries, _ := fs.ReadDir(bundled, "bundle/skills")
	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

func Source() string {
	b, _ := bundled.ReadFile("bundle/source.json")
	var source struct {
		Commit string `json:"commit"`
	}
	_ = json.Unmarshal(b, &source)
	return source.Commit
}

func Select(requested []string) ([]string, error) {
	known := map[string]bool{}
	for _, name := range Names() {
		known[name] = true
	}
	if len(requested) == 0 {
		for _, name := range Names() {
			if name != "flashdata-mcp" {
				requested = append(requested, name)
			}
		}
	}
	selected := map[string]bool{}
	for _, name := range requested {
		if !known[name] {
			return nil, fmt.Errorf("unknown skill %q; run flashdata skills list", name)
		}
		selected[name] = true
	}
	names := []string{}
	for name := range selected {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func digest(b []byte) string { value := sha256.Sum256(b); return hex.EncodeToString(value[:]) }

func files(name string) (map[string][]byte, error) {
	result := map[string][]byte{}
	root := "bundle/skills/" + name
	err := fs.WalkDir(bundled, root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		b, err := bundled.ReadFile(path)
		result[strings.TrimPrefix(path, root+"/")] = b
		return err
	})
	return result, err
}

// Check every existing ancestor; never follow a link into another skill or home.
func noLinks(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	for current := abs; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("skill destination contains a symbolic link: %s", current)
		}
		if parent := filepath.Dir(current); parent == current {
			break
		}
	}
	return nil
}

func relativeFile(root, name string) (string, error) {
	if !fs.ValidPath(name) || !filepath.IsLocal(filepath.FromSlash(name)) || strings.ContainsAny(name, "\\:") || name == "." {
		return "", fmt.Errorf("invalid installed skill path")
	}
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := noLinks(path); err != nil {
		return "", err
	}
	return path, nil
}

func readFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 4<<20 {
		return nil, fmt.Errorf("skill file is not a regular file within the size limit: %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, (4<<20)+1))
	if len(b) > 4<<20 {
		return nil, fmt.Errorf("skill file exceeds the size limit: %s", path)
	}
	return b, err
}

func readReceipt(root string) (Receipt, bool, error) {
	path, err := relativeFile(root, receiptName)
	if err != nil {
		return Receipt{}, false, err
	}
	b, err := readFile(path)
	if os.IsNotExist(err) {
		return Receipt{}, false, nil
	}
	if err != nil {
		return Receipt{}, false, err
	}
	var record Receipt
	if json.Unmarshal(b, &record) != nil || record.Version != 1 || record.Files == nil {
		return record, false, fmt.Errorf("invalid installation record in %s", root)
	}
	for file := range record.Files {
		if file == receiptName {
			return record, false, fmt.Errorf("invalid installation record in %s", root)
		}
		if _, err = relativeFile(root, file); err != nil {
			return record, false, err
		}
	}
	return record, true, nil
}

func checkSkill(root, name string) error {
	destination := filepath.Join(root, name)
	record, _, err := readReceipt(destination)
	if err != nil {
		return err
	}
	wanted, err := files(name)
	if err != nil {
		return err
	}
	all := map[string]bool{}
	for file := range record.Files {
		all[file] = true
	}
	for file := range wanted {
		all[file] = true
	}
	for file := range all {
		path, err := relativeFile(destination, file)
		if err != nil {
			return err
		}
		actual, err := readFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if next, ok := wanted[file]; ok && digest(actual) == digest(next) {
			continue
		}
		if prior, ok := record.Files[file]; ok && digest(actual) == prior {
			continue
		}
		return fmt.Errorf("existing skill file differs: %s; preserve or move your copy before installing", path)
	}
	return nil
}

func Check(roots, names []string) error {
	selected, err := Select(names)
	if err != nil {
		return err
	}
	for _, root := range roots {
		if err := noLinks(root); err != nil {
			return err
		}
		for _, name := range selected {
			if err := checkSkill(root, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeAtomic(path string, data []byte) error {
	if err := noLinks(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".flashdata-write-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if _, err = temp.Write(data); err == nil {
		err = temp.Sync()
	}
	closeErr := temp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Chmod(temp.Name(), 0644); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

func Install(roots, names []string) ([]Result, error) {
	selected, err := Select(names)
	if err != nil {
		return nil, err
	}
	if err = Check(roots, selected); err != nil {
		return nil, err
	}
	results := []Result{}
	for _, root := range roots {
		part, err := installRoot(root, selected)
		results = append(results, part...)
		if err != nil {
			return results, err
		}
	}
	return results, nil
}

func installRoot(root string, names []string) ([]Result, error) {
	if err := noLinks(root); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(root, ".flashdata-skills.lock")
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("skill installation is locked or not writable at %s; if a previous installer exited, remove its .flashdata-skills.lock and retry", root)
	}
	defer func() { _ = lock.Close(); _ = os.Remove(lockPath) }()
	// Recheck under the lock before changing any files in this destination.
	if err = Check([]string{root}, names); err != nil {
		return nil, err
	}
	results := []Result{}
	for _, name := range names {
		destination := filepath.Join(root, name)
		prior, managed, err := readReceipt(destination)
		if err != nil {
			return results, err
		}
		wanted, err := files(name)
		if err != nil {
			return results, err
		}
		record := Receipt{Version: 1, Source: Source(), Files: map[string]string{}}
		changed := false
		for file, data := range wanted {
			path, err := relativeFile(destination, file)
			if err != nil {
				return results, err
			}
			record.Files[file] = digest(data)
			actual, err := readFile(path)
			if err == nil && digest(actual) == record.Files[file] {
				continue
			}
			if err = writeAtomic(path, data); err != nil {
				return results, err
			}
			changed = true
		}
		for file := range prior.Files {
			if _, exists := wanted[file]; exists {
				continue
			}
			path, err := relativeFile(destination, file)
			if err != nil {
				return results, err
			}
			if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
				return results, err
			}
			changed = true
		}
		data, _ := json.MarshalIndent(record, "", "  ")
		if err = writeAtomic(filepath.Join(destination, receiptName), append(data, '\n')); err != nil {
			return results, err
		}
		status := "unchanged"
		if !managed {
			status = "installed"
		} else if changed {
			status = "updated"
		}
		results = append(results, Result{name, destination, status})
	}
	if err := Verify([]string{root}, names); err != nil {
		return results, err
	}
	return results, nil
}

func Verify(roots, names []string) error {
	selected, err := Select(names)
	if err != nil {
		return err
	}
	for _, root := range roots {
		for _, name := range selected {
			wanted, err := files(name)
			if err != nil {
				return err
			}
			for file, data := range wanted {
				path, err := relativeFile(filepath.Join(root, name), file)
				if err != nil {
					return err
				}
				actual, err := readFile(path)
				if err != nil {
					return err
				}
				if digest(actual) != digest(data) {
					return fmt.Errorf("skill verification failed: %s", path)
				}
			}
		}
	}
	return nil
}
