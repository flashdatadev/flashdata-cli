//go:build linux

package app

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultLinuxLoginOnHeadlessHost(t *testing.T) {
	if testing.Short() {
		t.Skip("build and real subprocess acceptance")
	}
	m, server := fixture(t)
	binary := buildBinary(t)
	root := filepath.Join(t.TempDir(), "state")
	t.Setenv("FLASHDATA_API_KEY", "")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/nonexistent/flashdata-test-bus")
	call := func(input string, args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		// Deliberately omit --credential-store: most acceptance helpers force file mode.
		base := []string{"--state-dir", root, "--json", "--non-interactive", "--data-origin", server.URL, "--management-origin", server.URL, "--web-origin", server.URL}
		cmd := exec.CommandContext(ctx, binary, append(base, args...)...)
		cmd.Env = cleanEnv()
		cmd.Stdin = strings.NewReader(input)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("%v failed: %v\n%s\n%s", args, err, stdout.String(), stderr.String())
		}
		if strings.Contains(stdout.String()+stderr.String(), "fd_owner_offline") {
			t.Fatal("API key leaked in output")
		}
		return stdout.String()
	}
	if out := call("fd_owner_offline\n", "auth", "login", "--api-key-stdin"); !strings.Contains(out, `"credentialProtection":"file"`) || !strings.Contains(out, `"status":"imported"`) {
		t.Fatal("login did not save private file credentials", out)
	}
	if out := call("", "auth", "status"); !strings.Contains(out, `"accessVerified":true`) {
		t.Fatal("next process could not reuse login", out)
	}
	if out := call("fd_owner_offline\n", "auth", "login", "--api-key-stdin"); !strings.Contains(out, `"status":"reused"`) {
		t.Fatal("repeated default login did not reuse credentials", out)
	}
	call("", "search", "headless login", "--run-id", "headless-search")
	if out := call("", "search", "headless login", "--run-id", "headless-search"); !strings.Contains(out, `"cached":true`) {
		t.Fatal("saved run was not reused", out)
	}
	if m.stats().QueryCount != 1 {
		t.Fatal("search was submitted more than once")
	}
	root = filepath.Join(t.TempDir(), "trial")
	if out := call("", "init", "--name", "Headless host"); !strings.Contains(out, `"credentialProtection":"file"`) || !strings.Contains(out, `"status":"trial"`) {
		t.Fatal("default initialization did not save credentials", out)
	}
	if out := call("", "auth", "status"); !strings.Contains(out, `"accessVerified":true`) {
		t.Fatal("next process could not reuse trial credentials", out)
	}
}
