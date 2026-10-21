package app

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	fd "github.com/flashdatadev/flashdata-go"
)

func TestEnvironmentBulkUpdatesPreserveSavedCredentials(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	_, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	s := stateStore(t, root)
	v, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	v.Phase = "installing"
	v.Credentials.HandoffProof = "fda_handoff_preserved"
	v.Credentials.SetupCredential = "fda_setup_preserved"
	v.RegistrationProof = "saved-registration-proof"
	original := v.Credentials
	if err = s.Save(context.Background(), v); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FLASHDATA_API_KEY", "fd_environment_offline")
	code, out, stderr := invoke(t, root, server.URL, `{"source":"youtube_metadata"}`, "bulk", "create", "--input", "-", "--run-id", "env-bulk")
	if code != 0 || strings.Contains(out, "fd_environment_offline") {
		t.Fatal(out, stderr)
	}
	v, err = s.Load(context.Background())
	if err != nil || v.Phase != "installing" || v.Credentials != original || v.RegistrationProof != "saved-registration-proof" || v.Uploads["bulk_offline1"] == nil {
		t.Fatal("session key destroyed saved profile", v, err)
	}
	t.Setenv("FLASHDATA_API_KEY", "")
	login(t, root, server.URL, original.APIKey)
	v, err = s.Load(context.Background())
	if err != nil || v.Phase != "installing" || v.Credentials != original || v.Uploads["bulk_offline1"] == nil {
		t.Fatal("same-key import destroyed recovery", v, err)
	}
}

func TestSourceSchemaCredentialDefinitionsAreNotRedacted(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	root := filepath.Join(t.TempDir(), "state")
	code, out, stderr := invoke(t, root, "", "", "sources", "show", "youtube_download")
	if code != 0 {
		t.Fatal(stderr)
	}
	var result struct {
		Data struct {
			Schema json.RawMessage `json:"inputSchema"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(err)
	}
	expected, err := fd.QuerySchema(fd.SourceYouTubeDownload)
	if err != nil {
		t.Fatal(err)
	}
	var left, right any
	_ = json.Unmarshal(expected, &left)
	_ = json.Unmarshal(result.Data.Schema, &right)
	rootSchema := right.(map[string]any)
	properties := rootSchema["properties"].(map[string]any)
	callback := properties["callback_url"].(map[string]any)
	if rootSchema["additionalProperties"] != false || callback["type"] != "string" {
		t.Fatal("schema does not describe runtime strictness/callbacks")
	}
	delete(rootSchema, "additionalProperties")
	delete(properties, "callback_url")
	if !reflect.DeepEqual(left, right) {
		t.Fatal("published input schema was corrupted by secret redaction")
	}
}

func TestBatchAccessVerificationIsConstantAndDryRunEstimates(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	before := m.stats()
	queries := make([]fd.Query, 20)
	for i := range queries {
		queries[i] = fd.Query{Source: fd.SourceGoogleSearch, Query: "same source"}
	}
	input, _ := json.Marshal(queries)
	code, _, stderr := invoke(t, root, server.URL, string(input), "batch", "submit", "--input", "-", "--run-id", "twenty")
	if code != 0 {
		t.Fatal(stderr)
	}
	after := m.stats()
	if after.AccessReads-before.AccessReads != 1 || after.IdentityReads-before.IdentityReads != 1 || after.QueryCount-before.QueryCount != 1 {
		t.Fatal("batch repeated management reads", after)
	}
	code, out, stderr := invoke(t, root, server.URL, string(input), "batch", "submit", "--input", "-", "--dry-run")
	if code != 0 || !strings.Contains(out, `"requestCount":20`) || !strings.Contains(out, `"estimates"`) || m.stats().QueryCount != after.QueryCount {
		t.Fatal(out, stderr)
	}
}

func TestRecurringPolicySurvivesFlagOmissionAndRecovery(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	code, _, stderr := invoke(t, root, server.URL, "", "search", "scheduled", "--run-id", "owner-only", "--recurring")
	if code != 0 {
		t.Fatal(stderr)
	}
	m.set(func(m *mockAPI) { m.NoRecurring = true })
	for _, args := range [][]string{{"search", "scheduled", "--run-id", "owner-only"}, {"runs", "recover", "owner-only"}, {"runs", "show", "owner-only", "--result"}} {
		code, out, stderr := invoke(t, root, server.URL, "", args...)
		if code != 1 || out != "" || !strings.Contains(stderr, "owned_access_required") {
			t.Fatal(args, out, stderr)
		}
	}
	if m.stats().QueryCount != 1 {
		t.Fatal("recurring run resubmitted")
	}
}

func TestOrphanedResultCannotTriggerAnotherPaidQuery(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	args := []string{"search", "orphan", "--run-id", "orphan"}
	if code, _, stderr := invoke(t, root, server.URL, "", args...); code != 0 {
		t.Fatal(stderr)
	}
	if err := os.Remove(stateStore(t, root).RunPath("orphan")); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := invoke(t, root, server.URL, "", args...)
	if code != 1 || !strings.Contains(stderr, "state_corrupt") || m.stats().QueryCount != 1 {
		t.Fatal(stderr)
	}
}

func TestBulkControlCommitsStageAndDiscardsSpentUploadCapability(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	_, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	if code, _, stderr := invoke(t, root, server.URL, `{"source":"youtube_metadata"}`, "bulk", "create", "--input", "-", "--run-id", "manual"); code != 0 {
		t.Fatal(stderr)
	}
	ids := filepath.Join(t.TempDir(), "ids.txt")
	if err := os.WriteFile(ids, []byte("dQw4w9WgXcQ\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"bulk", "upload", "bulk_offline1", "--input", ids}, {"bulk", "start", "bulk_offline1"}} {
		if code, _, stderr := invoke(t, root, server.URL, "", args...); code != 0 {
			t.Fatal(stderr)
		}
	}
	s := stateStore(t, root)
	v, err := s.Load(context.Background())
	if err != nil || len(v.Uploads) != 0 || len(v.BulkRuns) != 0 {
		t.Fatal("spent upload capability retained", v, err)
	}
	r, err := s.ReadRun("manual")
	if err != nil || r.Stage != "completed" || len(r.Input) == 0 || len(r.Receipt) == 0 {
		t.Fatal("manual stage/input/receipt lost", r, err)
	}
}

func TestUnusedFlagsAndAccidentalArgumentSecrets(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	root := filepath.Join(t.TempDir(), "state")
	for _, args := range [][]string{{"query", "--input", "-", "--wait"}, {"runs", "show", "r", "--output", "x.json"}, {"jobs", "get", "job_offline", "--output", "x.json"}, {"oauth-secret-accidentally-pasted"}} {
		code, out, stderr := invoke(t, root, "", `{"source":"google_search","query":"x"}`, args...)
		if code != 1 || out != "" || strings.Contains(stderr, "oauth-secret-accidentally-pasted") {
			t.Fatal(args, out, stderr)
		}
	}
}

func TestCredentialQueryInputRequiresPrivateFileOrStdin(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	input := `{"source":"youtube_download","video_id":"dQw4w9WgXcQ","params":{"storage":{"storage_type":"s3","storage_url":"test-bucket","access_key":"offline-access","secret_key":"offline-secret"}}}`
	path := filepath.Join(t.TempDir(), "shared-query.json")
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	if err := exposeFixture(path); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := invoke(t, root, server.URL, "", "query", "--input", path, "--mode", "async", "--dry-run")
	if code != 1 || !strings.Contains(stderr, "private_config_required") {
		t.Fatal(stderr)
	}
	code, out, stderr := invoke(t, root, server.URL, input, "query", "--input", "-", "--mode", "async", "--dry-run")
	if code != 0 || strings.Contains(out+stderr, "offline-access") || strings.Contains(out+stderr, "offline-secret") || m.stats().QueryCount != 0 {
		t.Fatal(out, stderr)
	}
}

func TestPendingRegistrationStatusAndWhitespaceScopeRecovery(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	m.set(func(m *mockAPI) { m.DropRegistration = true })
	args := []string{"init", "--name", "Pending", "--scopes", "google_search, youtube_metadata"}
	if code, _, _ := invoke(t, root, server.URL, "", args...); code != 1 {
		t.Fatal("expected lost registration response")
	}
	code, out, stderr := invoke(t, root, server.URL, "", "auth", "status")
	if code != 0 || !strings.Contains(out, `"phase":"registering"`) || !strings.Contains(out, `"nextCommand":"flashdata init"`) {
		t.Fatal(out, stderr)
	}
	if code, _, stderr = invoke(t, root, server.URL, "", args...); code != 0 {
		t.Fatal(stderr)
	}
	if m.stats().RegisterCount != 2 || m.stats().Proofs[0] != m.stats().Proofs[1] {
		t.Fatal("registration proof changed")
	}
}

func TestBulkJSONFieldsAreCaseSensitiveAndUpdateAcceptsStdin(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	_, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	for _, input := range []string{`{"Source":"youtube_metadata"}`, `{"source":"youtube_metadata","Options":{}}`} {
		if code, _, stderr := invoke(t, root, server.URL, input, "bulk", "create", "--input", "-"); code != 1 || !strings.Contains(stderr, "unsupported_field") {
			t.Fatal(stderr)
		}
	}
	code, _, stderr := invoke(t, root, server.URL, `{"options":{"include_formats":false},"callback_url":null}`, "bulk", "update", "bulk_offline1", "--input", "-")
	if code != 0 {
		t.Fatal(stderr)
	}
}

func TestEverySourceExampleUsesTheQueryWorkflow(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	_, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	for _, source := range fd.Sources() {
		t.Run(string(source), func(t *testing.T) {
			code, out, stderr := invoke(t, root, server.URL, "", "sources", "show", string(source))
			if code != 0 {
				t.Fatal(stderr)
			}
			var result struct {
				Data struct {
					Example json.RawMessage `json:"example"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(out), &result); err != nil {
				t.Fatal(err)
			}
			args := []string{"query", "--input", "-", "--run-id", "example-" + string(source)}
			if source == fd.SourceYouTubeDownload {
				args = append(args, "--mode", "async")
			}
			if code, _, stderr = invoke(t, root, server.URL, string(result.Data.Example), args...); code != 0 {
				t.Fatal("source example cannot be submitted", stderr)
			}
		})
	}
}

func TestRestrictedKeyCanReadPublicSchemaWithoutSubmitting(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_restricted")
	code, out, stderr := invoke(t, root, server.URL, "", "sources", "show", "google_search")
	if code != 0 || !strings.Contains(out, `"permitted":false`) || !strings.Contains(out, `"inputSchema"`) || m.stats().QueryCount != 0 {
		t.Fatal(out, stderr)
	}
}

func TestAcceptedBatchReturnsRemoteIDsWhenJournalCommitFails(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	path := stateStore(t, root).RunPath("commit-failure")
	m.set(func(m *mockAPI) {
		m.BeforeBatch = func() {
			// Keep the original intent while making atomic replacement fail.
			if err := os.Rename(path, path+".intent-backup"); err != nil {
				t.Error(err)
				return
			}
			if err := os.Mkdir(path, 0700); err != nil {
				t.Error(err)
			}
		}
	})
	input := `[{"source":"youtube_metadata","video_id":"dQw4w9WgXcQ"}]`
	args := []string{"batch", "submit", "--input", "-", "--run-id", "commit-failure"}
	code, out, stderr := invoke(t, root, server.URL, input, args...)
	if code != 1 || out != "" || !strings.Contains(stderr, `"remoteIds":["job_offline"]`) || !strings.Contains(stderr, "accepted_run_not_saved") {
		t.Fatal(out, stderr)
	}
	if code, _, _ := invoke(t, root, server.URL, input, args...); code != 1 || m.stats().QueryCount != 1 {
		t.Fatal("accepted batch was resubmitted")
	}
}

func TestBulkUploadUsesValidatedSnapshotWhenOriginalFileChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ids.txt")
	original := "dQw4w9WgXcQ\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	snapshot, digest, size, err := validateIDs(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeIDs(snapshot)
	if err = os.WriteFile(path, []byte("M7FIvfx5J10\n"), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(snapshot)
	if err != nil || string(data) != original || int64(len(data)) != size || digest == "" {
		t.Fatal("validated upload changed with original input", err)
	}
}

func TestExplicitImportIntoEnvironmentOnlyProfileRetainsRecovery(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "fd_environment_offline")
	_, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	code, _, stderr := invoke(t, root, server.URL, `{"source":"youtube_metadata"}`, "bulk", "create", "--input", "-", "--run-id", "env-only")
	if code != 0 {
		t.Fatal(stderr)
	}
	t.Setenv("FLASHDATA_API_KEY", "")
	login(t, root, server.URL, "fd_environment_offline")
	v, err := stateStore(t, root).Load(context.Background())
	if err != nil || v.Phase != "imported" || v.Credentials.APIKey != "fd_environment_offline" || v.Uploads["bulk_offline1"] == nil || v.BulkRuns["env-only"] != "bulk_offline1" {
		t.Fatal("explicit import discarded accepted-job recovery", v, err)
	}
}
