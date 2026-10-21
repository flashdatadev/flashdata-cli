package state

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	fd "github.com/flashdatadev/flashdata-go"
)

func newStore(t *testing.T, mode string) *Store {
	t.Helper()
	s, err := New(filepath.Join(t.TempDir(), "state"), "default", mode, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Ensure(); err != nil {
		t.Fatal(err)
	}
	return s
}
func TestVaultRoundTripAndNativeProtection(t *testing.T) {
	mode := "file"
	if runtime.GOOS == "windows" {
		mode = "auto"
	}
	s := newStore(t, mode)
	v := &Vault{Version: 1, Phase: "imported", Credentials: fd.AgentCredentials{APIKey: "fd_offline_vault_secret"}, RegistrationProof: "a-proof"}
	if err := s.Save(context.Background(), v); err != nil {
		t.Fatal(err)
	}
	out, err := s.Load(context.Background())
	if err != nil || out.Credentials.APIKey != v.Credentials.APIKey {
		t.Fatal("vault round trip", err)
	}
	raw, err := os.ReadFile(filepath.Join(s.Dir, "vault.json"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(v.Credentials.APIKey)) {
		t.Fatal("secret visible in envelope")
	}
	if runtime.GOOS == "windows" {
		var e envelope
		_ = json.Unmarshal(raw, &e)
		if e.Protection != "dpapi" {
			t.Fatal("Windows did not use DPAPI")
		}
		e.Data[0] ^= 1
		changed, _ := json.Marshal(e)
		if err = AtomicWrite(filepath.Join(s.Dir, "vault.json"), changed); err != nil {
			t.Fatal(err)
		}
		if _, err = s.Load(context.Background()); err == nil {
			t.Fatal("tampered DPAPI accepted")
		}
	}
}
func TestCorruptionAndMissingVaultNeverReset(t *testing.T) {
	s := newStore(t, "file")
	if err := s.Save(context.Background(), &Vault{Version: 1, Phase: "trial"}); err != nil {
		t.Fatal(err)
	}
	if err := AtomicWrite(filepath.Join(s.Dir, "vault.json"), []byte(`{"version":99}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(context.Background()); !errors.Is(err, ErrCorrupt) {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(s.Dir, "vault.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(context.Background()); !errors.Is(err, ErrMissing) {
		t.Fatal("missing tombstoned vault not recognized", err)
	}
}
func TestLocksExcludeAndRecoverAfterClose(t *testing.T) {
	s := newStore(t, "file")
	unlock, err := s.Lock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	if _, err = s.Lock(ctx); !errors.Is(err, ErrBusy) {
		t.Fatal("concurrent lock acquired", err)
	}
	unlock()
	again, err := s.Lock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	again()
	if _, err = os.Stat(filepath.Join(s.Dir, "profile.lock")); err != nil {
		t.Fatal("lock inode must remain stable", err)
	}
}
func TestRunAndStreamIntegrity(t *testing.T) {
	s := newStore(t, "file")
	r := &Run{ID: "daily:2026-10-22:Asia-Shanghai", Kind: "bulk:run", Status: "accepted", RemoteID: "bulk_offline1"}
	if err := s.SaveRun(r); err != nil {
		t.Fatal(err)
	}
	cause := errors.New("stream interrupted")
	err := s.SaveStream(r, func(w io.Writer) error { _, _ = io.WriteString(w, "{\"status\":\"completed\"}\n"); return cause })
	if !errors.Is(err, cause) || r.Status != "partial" {
		t.Fatal(err)
	}
	f, err := s.OpenResult(r)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(f)
	f.Close()
	if len(b) == 0 {
		t.Fatal("partial data lost")
	}
	if err = AtomicWrite(s.ResultPath(r.ID), []byte("tampered")); err != nil {
		t.Fatal(err)
	}
	if _, err = s.OpenResult(r); !errors.Is(err, ErrCorrupt) {
		t.Fatal("tampered result accepted", err)
	}
}
func TestProfilesAndRunPathsCannotEscape(t *testing.T) {
	for _, name := range []string{"../outside", "..", ".hidden", "", "a/b", "a\\b"} {
		if _, err := New(t.TempDir(), name, "file", false); err == nil {
			t.Fatal("unsafe profile accepted", name)
		}
	}
	s := newStore(t, "file")
	for _, id := range []string{"../x", "/absolute", "", "a/b"} {
		if _, err := s.ReadRun(id); err == nil {
			t.Fatal("unsafe run ID accepted")
		}
	}
	if filepath.Dir(s.RunPath("safe:id")) != filepath.Join(s.Dir, "runs") {
		t.Fatal("run path escaped")
	}
}
func TestEnvironmentKeyIsNotPersistedImplicitly(t *testing.T) {
	s := newStore(t, "file")
	v := &Vault{Phase: "environment", Credentials: fd.AgentCredentials{APIKey: "fd_environment_secret"}, BulkRuns: map[string]string{"r": "bulk_offline1"}}
	if err := s.Save(context.Background(), v); err != nil {
		t.Fatal(err)
	}
	saved, err := s.Load(context.Background())
	if err != nil || saved.Credentials.APIKey != "" || saved.BulkRuns["r"] == "" {
		t.Fatal("environment key persisted", err)
	}
	if v.Credentials.APIKey == "" {
		t.Fatal("caller mutated")
	}
}
func TestLegacyRunBarrierSurvivesMissingVault(t *testing.T) {
	s := newStore(t, "file")
	id := "stable-before-migration"
	barrier, _ := json.Marshal(map[string]any{Digest([]byte(id)): true})
	if err := AtomicWrite(filepath.Join(s.Dir, "legacy-runs.json"), barrier); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(context.Background()); !errors.Is(err, ErrMissing) {
		t.Fatal("migration barrier ignored", err)
	}
	if found, err := s.LegacyConflict(id); err != nil || !found {
		t.Fatal(found, err)
	}
}

func TestLargeJournalRecordsRoundTripAndListOnlyMetadata(t *testing.T) {
	s := newStore(t, "file")
	input, _ := json.Marshal(string(bytes.Repeat([]byte("x"), 2<<20)))
	r := &Run{ID: "large-batch", Kind: "batch", Status: "accepted", Input: input, Receipt: json.RawMessage(`{"id":"accepted"}`)}
	if err := s.SaveRun(r); err != nil {
		t.Fatal(err)
	}
	out, err := s.ReadRun(r.ID)
	if err != nil || !bytes.Equal(out.Input, r.Input) {
		t.Fatal("journal read limit is smaller than accepted input", err)
	}
	list, err := s.Runs()
	if err != nil || len(list) != 1 || len(list[0].Input) != 0 || len(list[0].Receipt) != 0 {
		t.Fatal("history listing retained full payloads", err)
	}
}

func TestOutputIsPrivateFromCreationAndCannotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private-output.json")
	f, err := CreateOutput(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := f.Stat()
	if err != nil || privateCheck(path, info) != nil {
		f.Close()
		t.Fatal("initial output permissions are unsafe", err)
	}
	_, err = f.Write([]byte("private result"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	if other, err := CreateOutput(path); err == nil {
		other.Close()
		t.Fatal("overwrote existing output")
	}
	b, err := ReadPrivate(path, 128)
	if err != nil || string(b) != "private result" {
		t.Fatal(err)
	}
}

func TestMigrationPinsSourceAndCannotResetUnrelatedHistory(t *testing.T) {
	s := newStore(t, "file")
	if err := s.StartMigration("source-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(context.Background()); !errors.Is(err, ErrMissing) {
		t.Fatal("incomplete migration did not block fresh access", err)
	}
	if err := s.StartMigration("source-a"); err != nil {
		t.Fatal("same-source recovery blocked", err)
	}
	if err := s.StartMigration("source-b"); err == nil {
		t.Fatal("different source replaced interrupted migration")
	}
	other := newStore(t, "file")
	if err := other.SaveRun(&Run{ID: "old", Status: "uncertain"}); err != nil {
		t.Fatal(err)
	}
	if err := other.StartMigration("source-a"); err == nil {
		t.Fatal("missing vault permitted history reset")
	}
}
