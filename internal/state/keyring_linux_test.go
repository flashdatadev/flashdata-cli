//go:build linux

package state

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	fd "github.com/flashdatadev/flashdata-go"
)

func TestSecretServiceNeverAutolaunchesOrUsesRemoteBus(t *testing.T) {
	for _, address := range []string{"", "autolaunch:", "tcp:host=example.com,port=1234", "unix:path=/tmp/flashdata-test-bus;autolaunch:"} {
		t.Setenv("DBUS_SESSION_BUS_ADDRESS", address)
		if _, err := keyringKey(context.Background(), "offline", true, false); !errors.Is(err, ErrVault) {
			t.Fatal("unexpected bus accepted", address, err)
		}
	}
}

func TestLinuxAutoStoreOnHeadlessHost(t *testing.T) {
	for _, address := range []string{"", "unix:path=/nonexistent/flashdata-test-bus", "autolaunch:", "tcp:host=example.com,port=1234"} {
		t.Run(address, func(t *testing.T) {
			t.Setenv("DBUS_SESSION_BUS_ADDRESS", address)
			assertAutoFileStore(t)
		})
	}
}

func assertAutoFileStore(t *testing.T) {
	t.Helper()
	s := newStore(t, "auto")
	v := &Vault{Phase: "imported", Credentials: fd.AgentCredentials{APIKey: "fd_headless_offline"}}
	if err := s.Save(context.Background(), v); err != nil {
		t.Fatal("new headless profile could not save", err)
	}
	if protection, err := s.Protection(); err != nil || protection != "file" {
		t.Fatal("expected file protection", protection, err)
	}
	for path, want := range map[string]os.FileMode{s.Dir: 0700, filepath.Join(s.Dir, "vault.json"): 0600} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != want || privateCheck(path, info) != nil {
			t.Fatal("incorrect private permissions", path, err)
		}
	}
	// Later commands must reuse the saved file even when their bus changes.
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/another/unavailable-bus")
	reopened, err := New(s.Root, s.Profile, "auto", false)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := reopened.Load(context.Background())
	if err != nil || saved == nil || saved.Credentials.APIKey != v.Credentials.APIKey {
		t.Fatal("default command could not reuse saved key", err)
	}
	if err = reopened.Save(context.Background(), saved); err != nil {
		t.Fatal("default command could not update saved profile", err)
	}
}

func TestLinuxAutoStoreWithSessionBusButNoSecretService(t *testing.T) {
	daemon, err := exec.LookPath("dbus-daemon")
	if err != nil {
		t.Skip("dbus-daemon is unavailable")
	}
	// Start an isolated local bus without service activation directories.
	config := filepath.Join(t.TempDir(), "bus.conf")
	if err = os.WriteFile(config, []byte(`<busconfig><type>session</type><listen>unix:tmpdir=/tmp</listen><auth>EXTERNAL</auth><policy context="default"><allow send_destination="*"/><allow receive_sender="*"/><allow own="*"/></policy></busconfig>`), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, daemon, "--nofork", "--print-address=1", "--config-file="+config)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	address, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatal("private bus did not start", err)
	}
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", strings.TrimSpace(address))
	assertAutoFileStore(t)
}

func TestLinuxAutoStoreDoesNotReplaceExistingProtection(t *testing.T) {
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/nonexistent/flashdata-test-bus")
	for _, existing := range []string{"keyring", "missing-vault", "run-history", "corrupt-vault", "empty-protection"} {
		t.Run(existing, func(t *testing.T) {
			s := newStore(t, "auto")
			var path string
			var before []byte
			switch existing {
			case "keyring", "corrupt-vault", "empty-protection":
				path = filepath.Join(s.Dir, "vault.json")
				before, _ = json.Marshal(envelope{Version: 1, Protection: "keyring", Data: []byte("saved encrypted credentials")})
				if existing == "corrupt-vault" {
					before = []byte("invalid vault")
				} else if existing == "empty-protection" {
					before = []byte(`{"version":1}`)
				}
			case "missing-vault":
				path = filepath.Join(s.Dir, "profile.json")
				before = []byte(`{"version":1,"initialized":true}`)
			case "run-history":
				path = s.RunPath("accepted")
				before = []byte(`{"id":"accepted","remoteId":"job_existing"}`)
			}
			if err := AtomicWrite(path, before); err != nil {
				t.Fatal(err)
			}
			if err := s.Save(context.Background(), &Vault{Phase: "imported"}); err == nil {
				t.Fatal("existing profile was replaced by fallback storage")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(after, before) {
				t.Fatal("existing profile was changed", err)
			}
			if existing == "missing-vault" || existing == "run-history" {
				if _, err := os.Stat(filepath.Join(s.Dir, "vault.json")); !os.IsNotExist(err) {
					t.Fatal("missing vault was recreated", err)
				}
			}
		})
	}
}

func TestLinuxAutoFallbackRespectsExplicitOSAndCancellation(t *testing.T) {
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/nonexistent/flashdata-test-bus")
	for _, mode := range []string{"os", "auto"} {
		t.Run(mode, func(t *testing.T) {
			s := newStore(t, mode)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "auto" {
				cancel()
			}
			if err := s.Save(ctx, &Vault{Phase: "imported"}); err == nil {
				t.Fatal("unexpected file fallback")
			}
			if s.Exists() {
				t.Fatal("failed save created a profile")
			}
		})
	}
}
