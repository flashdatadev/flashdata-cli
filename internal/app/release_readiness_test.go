package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestUncertainBatchKeepsCandidateIDsWithoutConfirmingAcceptance(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	m.set(func(m *mockAPI) { m.MalformedBatch = true })
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	input := `[{"source":"google_search","query":"offline"},{"source":"youtube_metadata","video_id":"dQw4w9WgXcQ"}]`
	args := []string{"batch", "submit", "--input", "-", "--run-id", "uncertain-batch"}
	code, out, stderr := invoke(t, root, server.URL, input, args...)
	if code != 1 || out != "" || !strings.Contains(stderr, `"candidateRemoteIds":["job_offline","job_offline1"]`) || !strings.Contains(stderr, `"acceptanceConfirmed":false`) {
		t.Fatal(out, stderr)
	}
	r, err := stateStore(t, root).ReadRun("uncertain-batch")
	if err != nil || r.Status != "uncertain" || len(r.RemoteIDs) != 0 || !reflect.DeepEqual(r.CandidateRemoteIDs, []string{"job_offline", "job_offline1"}) {
		t.Fatal("lost or promoted candidate IDs", r, err)
	}
	for _, command := range [][]string{args, {"runs", "recover", "uncertain-batch"}} {
		code, out, stderr = invoke(t, root, server.URL, input, command...)
		if code != 1 || out != "" || !strings.Contains(stderr, "candidateRemoteIds") || m.stats().QueryCount != 1 {
			t.Fatal("uncertain batch was replayed or treated as successful", out, stderr)
		}
	}
}

func TestPartialBulkExportNeverReportsCompleteSuccess(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	m.set(func(m *mockAPI) { m.TruncatePart = true })
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	ids := filepath.Join(t.TempDir(), "ids.txt")
	if err := os.WriteFile(ids, []byte("dQw4w9WgXcQ\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "partial.jsonl")
	code, out, stderr := invoke(t, root, server.URL, "", "bulk", "run", "--input", ids, "--run-id", "partial-stream", "--output", output)
	if code != 1 || out != "" || !strings.Contains(stderr, "partial_result") {
		t.Fatal("partial transfer reported success", out, stderr)
	}
	r, err := stateStore(t, root).ReadRun("partial-stream")
	if err != nil || r.Status != "partial" || r.ItemFailures != 0 || r.ResultHash == "" {
		t.Fatal("partial receipt journal lost", r, err)
	}
	data, err := os.ReadFile(output)
	if err != nil || !strings.Contains(string(data), `"status":"completed"`) {
		t.Fatal("available receipts not preserved", err)
	}
	code, out, stderr = invoke(t, root, server.URL, "", "runs", "show", "partial-stream", "--result")
	if code != 1 || out != "" || !strings.Contains(stderr, "partial_result") {
		t.Fatal("cached partial export reported success", out, stderr)
	}
	m.set(func(m *mockAPI) { m.TruncatePart = false })
	code, out, stderr = invoke(t, root, server.URL, "", "runs", "recover", "partial-stream")
	if code != 0 || !strings.Contains(out, `"status":"succeeded"`) {
		t.Fatal(out, stderr)
	}
	stats := m.stats()
	if stats.CreateCount != 1 || stats.UploadCount != 1 || stats.StartCount != 1 {
		t.Fatal("result recovery repeated paid work", stats)
	}
}

func TestStandaloneLostACKAndBulkStartRecoverAcrossProcesses(t *testing.T) {
	if testing.Short() {
		t.Skip("compiled binary and real fresh-process recovery")
	}
	m, server := fixture(t)
	binary := buildBinary(t)
	root := filepath.Join(t.TempDir(), "state")
	for _, args := range [][]string{{"init", "--name", "Recovery acceptance"}, {"auth", "claim"}} {
		if code, out, stderr := binaryCall(t, binary, root, server.URL, "", args...); code != 0 {
			t.Fatal(out, stderr)
		}
	}
	m.set(func(m *mockAPI) { m.Approve, m.DropACK = true, true })
	code, out, stderr := binaryCall(t, binary, root, "", "", "auth", "finish")
	if code != 1 || out != "" || !strings.Contains(stderr, "request_outcome_unknown") {
		t.Fatal("lost ACK was hidden", out, stderr)
	}
	s := stateStore(t, root)
	v, err := s.Load(context.Background())
	if err != nil || v.Phase != "installing" || v.Credentials.HandoffProof == "" || v.Credentials.APIKey != "fd_candidate_offline" {
		t.Fatal("candidate discarded after lost ACK", err)
	}
	if code, out, stderr = binaryCall(t, binary, root, "", "", "auth", "finish"); code != 0 {
		t.Fatal(out, stderr)
	}
	v, err = s.Load(context.Background())
	if err != nil || v.Phase != "installed" || v.Credentials.HandoffProof != "" || v.RegistrationProof != "" || m.stats().RegisterCount != 1 || m.stats().ACKCount != 2 {
		t.Fatal("installation recovery did not converge", err)
	}
	ids := filepath.Join(t.TempDir(), "ids.txt")
	if err = os.WriteFile(ids, []byte("dQw4w9WgXcQ\n"), 0600); err != nil {
		t.Fatal(err)
	}
	m.set(func(m *mockAPI) { m.DropStart = true })
	code, out, stderr = binaryCall(t, binary, root, "", "", "bulk", "run", "--input", ids, "--run-id", "bulk-recovery")
	if code != 1 || out != "" || !strings.Contains(stderr, "request_outcome_unknown") {
		t.Fatal("lost start was hidden", out, stderr)
	}
	v, err = s.Load(context.Background())
	if err != nil || v.Uploads["bulk_offline1"] == nil {
		t.Fatal("original upload evidence lost before reconciliation", err)
	}
	if code, out, stderr = binaryCall(t, binary, root, "", "", "runs", "recover", "bulk-recovery"); code != 0 {
		t.Fatal(out, stderr)
	}
	v, err = s.Load(context.Background())
	stats := m.stats()
	if err != nil || len(v.Uploads) != 0 || len(v.BulkRuns) != 0 || stats.CreateCount != 1 || stats.UploadCount != 1 || stats.StartCount != 1 {
		t.Fatal("recovery replayed a mutation or retained a spent capability", err, stats)
	}
}

func TestStandaloneEnvironmentKeyWorksAcrossProcesses(t *testing.T) {
	if testing.Short() {
		t.Skip("compiled binary and real fresh-process recovery")
	}
	m, server := fixture(t)
	binary := buildBinary(t)
	root := filepath.Join(t.TempDir(), "state")
	env := []string{"FLASHDATA_API_KEY=fd_environment_offline"}
	code, out, stderr := binaryCallWithEnv(t, binary, root, server.URL, "", env, "search", "offline", "--run-id", "env-process")
	if code != 0 {
		t.Fatal(out, stderr)
	}
	for _, args := range [][]string{
		{"account", "balance"},
		{"search", "offline", "--run-id", "env-process"},
		{"runs", "show", "env-process", "--result"},
	} {
		code, out, stderr = binaryCallWithEnv(t, binary, root, "", "", env, args...)
		if code != 0 {
			t.Fatal("environment key failed in a fresh process", out, stderr)
		}
	}
	code, out, stderr = binaryCall(t, binary, root, "", "", "account", "balance")
	if code != 1 || out != "" || !strings.Contains(stderr, "credentials_missing") {
		t.Fatal("environment key was persisted", out, stderr)
	}
	code, out, stderr = binaryCallWithEnv(t, binary, root, "", "", env, "account", "balance")
	if code != 0 || m.stats().QueryCount != 1 || m.stats().RegisterCount != 0 {
		t.Fatal("environment-key reuse failed or repeated work", out, stderr, m.stats())
	}
}
