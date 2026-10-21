package app

import (
	"context"
	"encoding/json"
	"github.com/flashdatadev/flashdata-cli/internal/state"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplicitMigrationPreservesSourceAndReplayBarrier(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	originalID := "legacy-original-id"
	key := state.Digest([]byte(originalID))
	legacy := map[string]any{"version": 2, "apiOrigin": s.URL, "dataOrigin": s.URL, "webOrigin": s.URL, "phase": "installed", "connectionId": "acn_offline1", "credentials": map[string]any{"connectionId": "acn_offline1", "apiKey": "fd_owner_offline", "generation": 2}, "operations": map[string]any{key: map[string]any{"hash": "old-input-digest", "state": "completed", "result": map[string]any{"source": "google_search", "status": "done", "results": []any{map[string]any{"title": "legacy preserved"}}, "usage": map[string]any{"credits": 0.02}}}}}
	plain, _ := json.Marshal(legacy)
	protection, data, err := legacyProtected(plain)
	if err != nil {
		t.Fatal(err)
	}
	envelope, _ := json.Marshal(map[string]any{"version": 1, "protection": protection, "data": data})
	legacyDir := filepath.Join(t.TempDir(), "legacy")
	legacyStore, err := state.New(legacyDir, "holder", "file", false)
	if err != nil {
		t.Fatal(err)
	}
	if err = legacyStore.Ensure(); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(legacyStore.Dir, "connection.json")
	if err = state.AtomicWrite(source, envelope); err != nil {
		t.Fatal(err)
	}
	code, out, stderr := invoke(t, root, s.URL, "", "auth", "migrate", "--from", source)
	if code != 0 {
		t.Fatal(out, stderr)
	}
	after, _ := os.ReadFile(source)
	if string(after) != string(envelope) {
		t.Fatal("legacy source changed")
	}
	code, out, stderr = invoke(t, root, s.URL, "", "runs", "show", "legacy-"+key, "--result")
	if code != 0 || !strings.Contains(out, "legacy preserved") {
		t.Fatal(out, stderr)
	}
	code, _, stderr = invoke(t, root, s.URL, `{"source":"google_search","query":"should not replay"}`, "query", "--input", "-", "--run-id", originalID)
	if code == 0 || !strings.Contains(stderr, "legacy_run_requires_inspection") {
		t.Fatal(stderr)
	}
	stats := m.stats()
	if stats.QueryCount != 0 || stats.RegisterCount != 0 {
		t.Fatal("migration dispatched paid work or a trial")
	}
	if code, _, stderr = invoke(t, root, s.URL, "", "auth", "logout"); code != 0 {
		t.Fatal(stderr)
	}
	if code, out, stderr = invoke(t, root, s.URL, "", "auth", "migrate", "--from", source); code != 0 || !strings.Contains(out, `"alreadyMigrated":true`) {
		t.Fatal(out, stderr)
	}
	v, err := stateStore(t, root).Load(context.Background())
	if err != nil || v.Phase != "logged_out" || v.Credentials.APIKey != "" {
		t.Fatal("repeated migration restored obsolete credentials", v, err)
	}
}
