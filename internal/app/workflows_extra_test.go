package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/flashdatadev/flashdata-cli/internal/state"
)

func TestSubtitleLanguageFlags(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	for _, action := range []string{"transcript", "captions"} {
		for _, tc := range []struct {
			flags []string
			want  any
		}{
			{nil, nil}, {[]string{"--language", "en"}, "en"},
			{[]string{"--language", "en,es"}, []any{"en", "es"}},
			{[]string{"--language", "en", "--language", "es"}, []any{"en", "es"}},
		} {
			for _, mode := range []string{"realtime", "async"} {
				args := append([]string{"youtube", action, "aircAruvnKk", "--dry-run", "--mode", mode}, tc.flags...)
				code, out, stderr := invoke(t, root, server.URL, "", args...)
				if code != 0 {
					t.Fatal(code, out, stderr)
				}
				var response struct {
					Data struct {
						Request struct{ Params map[string]any }
					}
				}
				if err := json.Unmarshal([]byte(out), &response); err != nil {
					t.Fatal(err)
				}
				params := response.Data.Request.Params
				if !reflect.DeepEqual(params["language"], tc.want) || params["auto"] != (action == "transcript") {
					t.Fatalf("wrong subtitle selection: %#v", params)
				}
				if tc.want == nil {
					if _, exists := params["language"]; exists {
						t.Fatal("omitted language was filled")
					}
				}
			}
		}
		for _, language := range []string{"", "en,en", "en,", "a,b,c,d,e,f,g,h,i,j,k"} {
			code, out, stderr := invoke(t, root, server.URL, "", "youtube", action, "aircAruvnKk", "--language", language, "--dry-run")
			if code == 0 {
				t.Fatal("invalid language accepted", out, stderr)
			}
		}
		input := `{"source":"youtube_` + action + `","video_id":"aircAruvnKk","params":{"language":["en"],"auto":false}}`
		code, out, stderr := invoke(t, root, server.URL, input, "query", "--input", "-", "--dry-run")
		if code != 0 || !strings.Contains(out, `"language":["en"]`) {
			t.Fatal("single-element array lost", out, stderr)
		}
		code, out, stderr = invoke(t, root, server.URL, "["+input+"]", "batch", "submit", "--input", "-", "--dry-run")
		if code != 0 || !strings.Contains(out, `"language":["en"]`) {
			t.Fatal("batch language array rejected", out, stderr)
		}
	}
	if m.stats().QueryCount != 0 || m.stats().CreateCount != 0 {
		t.Fatal("dry run dispatched paid work")
	}
}

func TestTimeoutIncludesWaitingForStdin(t *testing.T) {
	for _, args := range [][]string{{"query", "--input", "-"}, {"auth", "login", "--api-key-stdin"}} {
		t.Run(strings.Join(args[:2], "/"), func(t *testing.T) {
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()
			var out, stderr bytes.Buffer
			a := New(IO{In: reader, Out: &out, Err: &stderr}, "1.0.0", "test")
			started := time.Now()
			code := a.Execute(context.Background(), append([]string{"--json", "--non-interactive", "--timeout", "20ms"}, args...))
			if code != 1 || !strings.Contains(stderr.String(), `"code":"timeout"`) || out.Len() != 0 || time.Since(started) > 2*time.Second {
				t.Fatal("stdin deadline ignored", code, out.String(), stderr.String())
			}
		})
	}
}

func TestExistingOutputPreventsBatchAndBulkCreation(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	output := filepath.Join(t.TempDir(), "existing.json")
	if err := os.WriteFile(output, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		args  []string
		input string
	}{
		{[]string{"batch", "submit", "--input", "-", "--run-id", "no-batch", "--output", output}, `[{"source":"youtube_metadata","video_id":"dQw4w9WgXcQ"}]`},
		{[]string{"bulk", "create", "--input", "-", "--run-id", "no-bulk", "--output", output}, `{"source":"youtube_metadata"}`},
	} {
		code, _, stderr := invoke(t, root, server.URL, test.input, test.args...)
		if code != 1 || !strings.Contains(stderr, "output_exists") {
			t.Fatal(stderr)
		}
	}
	if m.stats().QueryCount != 0 || m.stats().CreateCount != 0 {
		t.Fatal("submitted before output preflight")
	}
	data, err := os.ReadFile(output)
	if err != nil || string(data) != "original" {
		t.Fatal("existing output changed", err)
	}
}

func TestBatchWaitResumesAcceptedRunWithoutResubmission(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	input := `[{"source":"youtube_metadata","video_id":"dQw4w9WgXcQ"}]`
	args := []string{"batch", "submit", "--input", "-", "--run-id", "batch-recover"}
	if code, _, stderr := invoke(t, root, server.URL, input, args...); code != 0 {
		t.Fatal(stderr)
	}
	if code, _, stderr := invoke(t, root, server.URL, input, append(args, "--wait")...); code != 0 {
		t.Fatal(stderr)
	}
	run, err := stateStore(t, root).ReadRun("batch-recover")
	if err != nil || run == nil || run.Status != "succeeded" || m.stats().QueryCount != 1 {
		t.Fatal("accepted batch was resubmitted or not recovered", run, err)
	}
}

func TestProofSavedBeforeRegistrationAndReusedAfterLostResponse(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	m.set(func(m *mockAPI) { m.DropRegistration = true })
	m.BeforeRegistration = func() {
		v, err := stateStore(t, root).Load(context.Background())
		if err != nil || v == nil || v.Phase != "registering" || v.RegistrationProof == "" || v.Registration == nil {
			t.Error("registration dispatched before proof/input commit", err)
		}
	}
	code, _, err := invoke(t, root, s.URL, "", "init", "--name", "Persisted proof")
	if code == 0 || !strings.Contains(err, "request_outcome_unknown") {
		t.Fatal(err)
	}
	code, _, err = invoke(t, root, s.URL, "", "init")
	if code != 0 {
		t.Fatal(err)
	}
	if m.stats().RegisterCount != 2 || len(m.stats().Proofs) != 2 || m.stats().Proofs[0] != m.stats().Proofs[1] {
		t.Fatal("registration proof changed")
	}
}
func TestFailedFreshVerificationKeepsCandidateAndNeverACKs(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	code, _, err := invoke(t, root, s.URL, "", "init", "--name", "Candidate")
	if code != 0 {
		t.Fatal(err)
	}
	m.set(func(m *mockAPI) { m.Approve = true })
	var out, stderr bytes.Buffer
	a := New(IO{In: strings.NewReader(""), Out: &out, Err: &stderr}, "1.0.0", "test")
	a.VerifyFresh = func(context.Context, *state.Store, *state.Vault) error {
		return errors.New("simulated verification failure")
	}
	code = a.Execute(context.Background(), []string{"--state-dir", root, "--credential-store", "file", "--json", "--non-interactive", "auth", "finish"})
	if code == 0 || m.stats().ACKCount != 0 {
		t.Fatal("ACK before verification", stderr.String())
	}
	v, e := stateStore(t, root).Load(context.Background())
	if e != nil || v.Phase != "installing" || v.Credentials.APIKey != "fd_candidate_offline" || v.Credentials.HandoffProof == "" {
		t.Fatal("candidate/proof lost", e)
	}
}
func TestCommandCoverageAndBulkManualRecovery(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	_, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, s.URL, "fd_owner_offline")
	commands := [][]string{{"discover"}, {"auth", "status"}, {"sources", "list"}, {"sources", "show", "google_search"}, {"sources", "show", "youtube_download"}, {"account", "balance"}, {"account", "pricing", "--source", "google_search", "--request-count", "3"}, {"account", "limits"}, {"account", "usage", "--source", "youtube_metadata", "--from", "2026-10-01T00:00:00Z", "--to", "2026-10-02T00:00:00Z"}, {"doctor"}, {"doctor", "--offline"}, {"youtube", "search", "tutorials", "--max-results", "2"}, {"youtube", "metadata", "dQw4w9WgXcQ"}, {"youtube", "transcript", "dQw4w9WgXcQ"}, {"youtube", "captions", "dQw4w9WgXcQ"}, {"jobs", "get", "job_offline"}, {"jobs", "wait", "job_offline"}, {"jobs", "result", "job_offline"}, {"bulk", "list"}, {"runs", "list"}}
	for _, args := range commands {
		t.Run(strings.Join(args[:min(2, len(args))], "/"), func(t *testing.T) {
			code, out, err := invoke(t, root, s.URL, "", args...)
			if code != 0 || !strings.Contains(out, `"schemaVersion":1`) {
				t.Fatal(args, out, err)
			}
		})
	}
	code, _, err := invoke(t, root, s.URL, `[{"source":"google_search","query":"batch"},{"source":"youtube_metadata","video_id":"dQw4w9WgXcQ"}]`, "batch", "submit", "--input", "-", "--run-id", "batch", "--wait")
	if code != 0 {
		t.Fatal(err)
	}
	store := stateStore(t, root)
	config := filepath.Join(store.Dir, "bulk-config.json")
	if e := state.AtomicWrite(config, []byte(`{"source":"youtube_metadata","options":{"include_formats":false}}`)); e != nil {
		t.Fatal(e)
	}
	code, _, err = invoke(t, root, s.URL, "", "bulk", "create", "--input", config, "--run-id", "manual")
	if code != 0 {
		t.Fatal(err)
	}
	code, _, err = invoke(t, root, s.URL, "", "runs", "recover", "manual")
	if code == 0 || !strings.Contains(err, "bulk_input_required") {
		t.Fatal(err)
	}
	ids := filepath.Join(t.TempDir(), "ids.txt")
	_ = os.WriteFile(ids, []byte("dQw4w9WgXcQ\n"), 0600)
	for _, args := range [][]string{{"bulk", "upload", "bulk_offline1", "--input", ids, "--run-id", "manual"}, {"bulk", "start", "bulk_offline1"}, {"bulk", "get", "bulk_offline1"}, {"bulk", "pause", "bulk_offline1"}, {"bulk", "resume", "bulk_offline1"}, {"bulk", "wait", "bulk_offline1"}, {"runs", "recover", "manual"}, {"runs", "show", "manual", "--result"}, {"bulk", "results", "bulk_offline1", "--run-id", "read-again"}, {"bulk", "cancel", "bulk_offline1"}} {
		code, _, err = invoke(t, root, s.URL, "", args...)
		if code != 0 {
			t.Fatal(args, err)
		}
	}
}
func TestDryRunRedactionAndDuplicateFields(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, s.URL, "fd_owner_offline")
	input := `{"source":"youtube_download","video_id":"dQw4w9WgXcQ","params":{"storage":{"storage_type":"s3","storage_url":"bucket","access_key":"private-access","secret_key":"private-secret"}}}`
	code, out, err := invoke(t, root, s.URL, input, "query", "--input", "-", "--mode", "async", "--dry-run")
	if code != 0 || strings.Contains(out, "private-access") || strings.Contains(out, "private-secret") {
		t.Fatal(out, err)
	}
	for _, bad := range []string{`{"source":"google_search","source":"youtube_search","query":"x"}`, `{"source":"google_search","query":"x","params":{"num":5,"num":10}}`, `{} {}`} {
		code, _, err = invoke(t, root, s.URL, bad, "query", "--input", "-")
		if code == 0 {
			t.Fatal("duplicate/multiple values accepted", err)
		}
	}
	if m.stats().QueryCount != 0 || m.stats().CreateCount != 0 {
		t.Fatal("dry-run/invalid input dispatched")
	}
}
