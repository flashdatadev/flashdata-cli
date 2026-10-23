package app

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	// Release verification reuses the same acceptance scenarios against the
	// executable extracted from the candidate archive, with mock APIs only.
	if candidate := os.Getenv("FLASHDATA_TEST_BINARY"); candidate != "" {
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(absolute)
		if err != nil || !info.Mode().IsRegular() {
			t.Fatal("invalid packaged test executable", err)
		}
		return absolute
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	name := "flashdata"
	goExe := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		name += ".exe"
		goExe += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, goExe, "build", "-o", binary, "./cmd/flashdata")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return binary
}
func cleanEnv() []string {
	env := []string{}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(e), "FLASHDATA_") {
			env = append(env, e)
		}
	}
	return env
}
func binaryCall(t *testing.T, binary, root, url, input string, args ...string) (int, string, string) {
	return binaryCallWithEnv(t, binary, root, url, input, nil, args...)
}
func binaryCallWithEnv(t *testing.T, binary, root, url, input string, env []string, args ...string) (int, string, string) {
	t.Helper()
	base := []string{"--state-dir", root, "--json", "--non-interactive"}
	if runtime.GOOS != "windows" {
		base = append(base, "--credential-store", "file")
	}
	if url != "" {
		base = append(base, "--data-origin", url, "--management-origin", url, "--web-origin", url)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, append(base, args...)...)
	cmd.Env = append(cleanEnv(), env...)
	cmd.Stdin = strings.NewReader(input)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	configureChild(cmd)
	if err := cmd.Run(); err != nil {
		return 1, out.String(), stderr.String()
	}
	return 0, out.String(), stderr.String()
}

func TestStandaloneBinaryFreshProcessAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip("build and real subprocess acceptance")
	}
	m, s := fixture(t)
	binary := buildBinary(t)
	root := filepath.Join(t.TempDir(), "state")
	code, out, err := binaryCall(t, binary, root, s.URL, "", "init", "--name", "Fresh machine", "--scopes", "google_search,youtube_metadata")
	if code != 0 {
		t.Fatal(out, err)
	}
	code, out, err = binaryCall(t, binary, root, "", "", "search", "first query", "--run-id", "first")
	if code != 0 {
		t.Fatal(out, err)
	}
	code, out, err = binaryCall(t, binary, root, "", "", "auth", "claim")
	if code != 0 || !strings.Contains(out, "confirmationCode") {
		t.Fatal(out, err)
	}
	m.mu.Lock()
	m.Approve = true
	m.mu.Unlock()
	code, out, err = binaryCall(t, binary, root, "", "", "auth", "finish", "--wait")
	if code != 0 {
		t.Fatal("real fresh-process verification failed", out, err)
	}
	m.mu.Lock()
	ack := m.ACKCount
	m.mu.Unlock()
	if ack != 1 {
		t.Fatal("ACK was not made after independent verification")
	}
	for i := 0; i < 2; i++ {
		code, out, err = binaryCall(t, binary, root, "", "", "search", "scheduled", "--recurring", "--run-id", "daily:2026-10-22:Asia-Shanghai")
		if code != 0 {
			t.Fatal(out, err)
		}
	}
	m.mu.Lock()
	calls := m.QueryCount
	m.Revoked = true
	m.mu.Unlock()
	if calls != 2 {
		t.Fatal("duplicate submission from another process", calls)
	}
	code, out, err = binaryCall(t, binary, root, "", "", "search", "scheduled", "--recurring", "--run-id", "daily:2026-10-22:Asia-Shanghai")
	if code == 0 || out != "" {
		t.Fatal("revoked cached data was served", err)
	}
	code, out, err = binaryCall(t, binary, root, "", "", "--version")
	if code != 0 || !strings.Contains(out, "1.1.2") {
		t.Fatal(out, err)
	}
}

// A helper process holds the OS lock, then is killed by the parent to verify
// kernel release without deleting/reclaiming a potentially live lock inode.
func TestLockHelperProcess(t *testing.T) {
	if os.Getenv("FD_TEST_LOCK_HELPER") == "" {
		return
	}
	root := os.Getenv("FD_TEST_STATE")
	s := stateStore(t, root)
	release, err := s.Lock(context.Background())
	if err != nil {
		os.Exit(4)
	}
	defer release()
	_, _ = os.Stdout.WriteString("LOCKED\n")
	for {
		time.Sleep(time.Hour)
	}
}
func TestProcessLockReleasedAfterCrash(t *testing.T) {
	if testing.Short() {
		t.Skip("process crash regression")
	}
	root := filepath.Join(t.TempDir(), "state")
	s := stateStore(t, root)
	if err := s.Ensure(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLockHelperProcess$")
	cmd.Env = append(cleanEnv(), "FD_TEST_LOCK_HELPER=1", "FD_TEST_STATE="+root)
	configureChild(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	line := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		if scanner.Scan() {
			line <- scanner.Text()
		} else {
			line <- ""
		}
	}()
	select {
	case value := <-line:
		if value != "LOCKED" {
			t.Fatal("helper did not acquire lock", value)
		}
	case <-ctx.Done():
		t.Fatal("helper timeout")
	}
	lockCtx, stop := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer stop()
	if unlockUnexpected, e := s.Lock(lockCtx); e == nil {
		unlockUnexpected()
		t.Fatal("another process acquired the lock")
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	cmd.Process = nil
	unlock, err := s.Lock(context.Background())
	if err != nil {
		t.Fatal("crash left stale lock", err)
	}
	unlock()
}
