package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/flashdatadev/flashdata-cli/internal/state"
)

func TestSTSCredentialsReachAPIAndStayOutOfLocalOutput(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	for _, provider := range []string{"s3", "s3_compatible", "oss", "cos", "bos", "tos"} {
		t.Run(provider, func(t *testing.T) {
			mock, _ := fixture(t)
			var mu sync.Mutex
			var bodies []map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost || r.Method == http.MethodPatch {
					body, _ := io.ReadAll(r.Body)
					r.Body = io.NopCloser(bytes.NewReader(body))
					var value map[string]any
					if json.Unmarshal(body, &value) == nil {
						mu.Lock()
						bodies = append(bodies, value)
						mu.Unlock()
					}
				}
				mock.handler(w, r)
			}))
			t.Cleanup(server.Close)
			mock.set(func(m *mockAPI) { m.URL = server.URL })
			root := filepath.Join(t.TempDir(), "state")
			login(t, root, server.URL, "fd_owner_offline")
			storage := map[string]any{"storage_type": provider, "storage_url": "test-bucket", "region": "us-east-1", "endpoint_url": "https://storage.example.com", "access_key": "temporary-access", "secret_key": "temporary-secret", "session_token": "temporary-session-credential"}
			encoded, _ := json.Marshal(storage)
			path := filepath.Join(t.TempDir(), "private", "storage.json")
			if err := state.AtomicWrite(path, encoded); err != nil {
				t.Fatal(err)
			}
			check := func(input string, args ...string) {
				t.Helper()
				code, out, stderr := invoke(t, root, server.URL, input, args...)
				if code != 0 || strings.Contains(out+stderr, "temporary-") {
					t.Fatal(args, code, out, stderr)
				}
			}
			check("", "youtube", "download", "dQw4w9WgXcQ", "--storage-config", path, "--dry-run")
			check("", "youtube", "download", "dQw4w9WgXcQ", "--storage-config", path, "--run-id", "sts-download")
			for _, source := range []string{"youtube_download", "youtube_metadata"} {
				field := "result_storage"
				if source == "youtube_download" {
					field = "storage"
				}
				input, _ := json.Marshal(map[string]any{"source": source, "options": map[string]any{field: storage}})
				check(string(input), "bulk", "create", "--input", "-", "--run-id", "sts-"+source)
			}
			patch, _ := json.Marshal(map[string]any{"options": map[string]any{"result_storage": storage}})
			check(string(patch), "bulk", "update", "bulk_offline1", "--input", "-")
			mu.Lock()
			defer mu.Unlock()
			if len(bodies) != 4 {
				t.Fatalf("expected download, two bulk creations and credential update; got %d", len(bodies))
			}
			for _, body := range bodies {
				data, _ := json.Marshal(body)
				if !bytes.Contains(data, []byte(`"session_token":"temporary-session-credential"`)) {
					t.Fatal("API did not receive the original session token")
				}
			}
			for _, id := range []string{"sts-download", "sts-youtube_download", "sts-youtube_metadata"} {
				run, err := stateStore(t, root).ReadRun(id)
				if err != nil {
					t.Fatal(err)
				}
				data, _ := json.Marshal(run)
				if bytes.Contains(data, []byte("temporary-")) {
					t.Fatal("saved run contains storage credentials")
				}
			}
		})
	}
}

func TestSessionTokenRedactionAndPrivateInput(t *testing.T) {
	value := map[string]any{"session_token": "raw-session-value"}
	a := New(IO{}, "test", "test")
	a.rememberSecrets(value)
	encoded, _ := json.Marshal(a.redacted(map[string]any{"sessionToken": "unknown-secret", "error": "rejected raw-session-value"}))
	if bytes.Contains(encoded, []byte("unknown-secret")) || bytes.Contains(encoded, []byte("raw-session-value")) {
		t.Fatal("structured or echoed session token leaked")
	}
	text := a.redactedResponseText("Session token: unknown-secret\nraw-session-value\nhttps://storage.example/?X-Amz-Security-Token=another-secret")
	for _, secret := range []string{"unknown-secret", "raw-session-value", "another-secret"} {
		if strings.Contains(text, secret) {
			t.Fatal("formatted output leaked session token")
		}
	}
	path := filepath.Join(t.TempDir(), "private", "shared.json")
	if err := state.AtomicWrite(path, []byte(`{"session_token":"raw-session-value"}`)); err != nil {
		t.Fatal(err)
	}
	if err := exposeFixture(path); err != nil {
		t.Fatal(err)
	}
	if err := requirePrivateInput(path, value); err == nil {
		t.Fatal("session-token-only input did not require a private file")
	}
}
