package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
)

type mockAPI struct {
	AccessReads, IdentityReads                       int
	NoRecurring                                      bool
	DropRegistration                                 bool
	DropACK, DropStart, TruncatePart, MalformedBatch bool
	BeforeRegistration                               func()
	BeforeBatch                                      func()

	mu                                                                        sync.Mutex
	URL                                                                       string
	QueryCount, CreateCount, UploadCount, StartCount, RegisterCount, ACKCount int
	Proofs                                                                    []string
	Revoked, Approve, Installed, DropQuery, IdentityMissing, Partial, NoTrial bool
	BulkSource                                                                string
	BulkStatus                                                                string
	Jobs                                                                      map[string]fd.Source
}

func send(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
func fixture(t *testing.T) (*mockAPI, *httptest.Server) {
	t.Helper()
	m := &mockAPI{BulkSource: "youtube_metadata", BulkStatus: "awaiting_upload"}
	s := httptest.NewServer(http.HandlerFunc(m.handler))
	m.mu.Lock()
	m.URL = s.URL
	m.mu.Unlock()
	t.Cleanup(s.Close)
	return m, s
}
func (m *mockAPI) handler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := r.Header.Get("X-API-Key")
	sourceList := fd.Sources()
	if strings.HasPrefix(r.URL.Path, "/v2/agent-connections") {
		if key != "" {
			w.WriteHeader(400)
			send(w, map[string]string{"code": "agent_ambiguous_credentials"})
			return
		}
	} else if strings.HasPrefix(r.URL.Path, "/v2/api-access") || strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/v2/jobs") || strings.HasPrefix(r.URL.Path, "/v2/bulk-jobs") {
		if key == "" || m.Revoked {
			w.WriteHeader(401)
			send(w, map[string]string{"code": "invalid_api_key", "detail": "Key unavailable"})
			return
		}
	}
	if strings.HasPrefix(r.URL.Path, "/signed/") {
		if key != "" || r.Header.Get("Authorization") != "" {
			panic("signed request leaked credentials")
		}
		if strings.Contains(r.URL.Path, "part") {
			var b bytes.Buffer
			g := gzip.NewWriter(&b)
			_, _ = io.WriteString(g, "{\"index\":0,\"status\":\"completed\",\"video_id\":\"dQw4w9WgXcQ\",\"results\":[],\"future\":9007199254740993123}\n")
			if m.Partial {
				_, _ = io.WriteString(g, "{\"index\":1,\"status\":\"invalid\"}\n")
			}
			_ = g.Close()
			part := b.Bytes()
			if m.TruncatePart {
				part = part[:len(part)-8]
			}
			_, _ = w.Write(part)
			return
		}
		send(w, map[string]any{"source": "youtube_metadata", "status": "done", "results": []any{map[string]any{"title": "test", "n": int64(9007199254740993)}}})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/v2/jobs/job_offline") {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v2/jobs/"), "/results")
		if strings.HasSuffix(r.URL.Path, "/results") {
			w.Header().Set("Location", m.URL+"/signed/"+id)
			w.WriteHeader(302)
			return
		}
		source := m.Jobs[id]
		if source == "" {
			source = fd.SourceYouTubeMetadata
		}
		send(w, map[string]any{"job": map[string]any{"id": id, "source": source, "status": "completed"}})
		return
	}
	switch r.URL.Path {
	case "/v2/health":
		send(w, map[string]any{"status": "healthy", "service": "flashdata-api", "version": "test"})
	case "/v2/plans":
		if key != "" {
			w.WriteHeader(400)
			return
		}
		send(w, map[string]any{"currency": "USD", "data": []any{}})
	case "/v2/status":
		if key != "" {
			w.WriteHeader(400)
			return
		}
		send(w, map[string]any{"status": "operational"})
	case "/v2/catalog/sources":
		data := []any{}
		for _, s := range sourceList {
			data = append(data, map[string]any{"id": s, "name": s, "modes": []string{"realtime", "async", "batch"}})
		}
		send(w, map[string]any{"data": data})
	case "/v2/api-access":
		m.AccessReads++
		if key == "fd_restricted" {
			sourceList = []fd.Source{fd.SourceYouTubeMetadata}
		}
		send(w, map[string]any{"sources": sourceList, "result_origin": m.URL})
	case "/v2/api-access/identity":
		m.IdentityReads++
		if m.IdentityMissing {
			w.WriteHeader(404)
			send(w, map[string]string{"code": "not_found"})
			return
		}
		owned := key != "fd_trial_offline"
		installed := owned && !m.NoRecurring
		if key == "fd_candidate_offline" {
			installed = m.Installed
		}
		ownership, account, accountType := "owned", "acct_offline", "owned"
		if !owned {
			ownership, account, accountType = "unowned", "ata_offline", "temporary"
		}
		var connectionID *string
		if !owned || key == "fd_candidate_offline" {
			connectionID = fd.Ptr("acn_offline1")
		}
		send(w, fd.APIKeyIdentity{APIKeyID: "key_offline", AccountID: account, AccountType: accountType, Ownership: ownership, Access: "ready", Installed: installed, RecurringAllowed: owned && installed, ConnectionID: connectionID, Scopes: sourceList, VerifiedAt: "2026-10-21T00:00:00Z"})
	case "/v2/api-access/balance":
		send(w, map[string]any{"balanceCredits": 1, "availableCredits": 1, "unit": "Credits", "scope": "account"})
	case "/v2/api-access/pricing":
		send(w, map[string]any{"unit": "Credits", "requestCount": 1, "prices": []any{map[string]any{"source": r.URL.Query().Get("source"), "available": true, "creditsPerRequest": 0.02, "estimatedCredits": 0.02}}})
	case "/v2/api-access/limits":
		send(w, map[string]any{"limits": map[string]any{"concurrency": 2, "batchSize": 20, "pendingJobs": 100}})
	case "/v2/usage":
		send(w, map[string]any{"unit": "Credits", "totals": map[string]any{"requests": 1, "credits": 0.02}})
	case "/v1/queries/realtime", "/v1/queries", "/v1/queries/batch":
		m.QueryCount++
		if m.DropQuery {
			conn, _, _ := w.(http.Hijacker).Hijack()
			_ = conn.Close()
			return
		}
		var q fd.Query
		if r.URL.Path == "/v1/queries/batch" {
			if m.BeforeBatch != nil {
				m.BeforeBatch()
			}
			var b struct{ Queries []fd.Query }
			_ = json.NewDecoder(r.Body).Decode(&b)
			jobs := []any{}
			if m.Jobs == nil {
				m.Jobs = make(map[string]fd.Source)
			}
			for i, query := range b.Queries {
				id := "job_offline"
				if i > 0 {
					id = fmt.Sprintf("job_offline%d", i)
				}
				m.Jobs[id] = query.Source
				jobs = append(jobs, map[string]any{"id": id, "source": query.Source, "status": "pending"})
			}
			if m.MalformedBatch && len(jobs) > 0 {
				jobs[0].(map[string]any)["source"] = "inconsistent_source"
			}
			send(w, map[string]any{"queries": jobs, "usage": map[string]any{"credits": 0.04}})
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&q)
		if r.URL.Path == "/v1/queries" {
			if m.Jobs == nil {
				m.Jobs = map[string]fd.Source{}
			}
			m.Jobs["job_offline"] = q.Source
			send(w, map[string]any{"id": "job_offline", "source": q.Source, "status": "pending", "usage": map[string]any{"credits": 0.02}})
			return
		}
		send(w, map[string]any{"source": q.Source, "status": "done", "created_at": "2026-10-21T00:00:00Z", "results": []any{map[string]any{"organic": []any{map[string]any{"title": "result", "link": "https://example.com/?q=provenance"}}, "n": int64(9007199254740993)}}, "usage": map[string]any{"credits": 0.02}})
	case "/v2/agent-connections/capabilities":
		send(w, fd.OnboardingCapabilities{ProtocolVersion: 2, Available: !m.NoTrial, DataOrigin: m.URL, Trial: fd.OnboardingTrial{Credits: 1}})
	case "/v2/agent-connections/registrations":
		m.RegisterCount++
		m.Proofs = append(m.Proofs, r.Header.Get("Authorization"))
		if m.BeforeRegistration != nil {
			m.BeforeRegistration()
		}
		if m.DropRegistration {
			m.DropRegistration = false
			conn, _, _ := w.(http.Hijacker).Hijack()
			_ = conn.Close()
			return
		}
		var input fd.RegistrationRequest
		_ = json.NewDecoder(r.Body).Decode(&input)
		send(w, map[string]any{"connection": fd.AgentConnection{ID: "acn_offline1", Name: input.Name, Ownership: "unowned", Lifecycle: "valid", Access: "ready", Scopes: input.Scopes, Generation: 1, TrialCredits: 1, TrialRequestCounts: json.RawMessage(`{"google_search":50}`)}, "credentials": fd.AgentCredentials{ConnectionID: "acn_offline1", APIKey: "fd_trial_offline", SetupCredential: "fda_setup_offline", HandoffProof: "fda_handoff_offline", Generation: 1}, "dataOrigin": m.URL, "deniedScopes": []any{}})
	case "/v2/agent-connections/acn_offline1/claims":
		send(w, fd.AgentClaim{ClaimID: "acl_offline1", ClaimPath: "/connect/claim/acl_offline1", ConfirmationCode: "12345678", ExpiresAt: "2026-10-22T00:00:00Z", PollAfterSeconds: 1})
	case "/v2/agent-connections/acn_offline1/onboarding":
		status := "awaiting_claim"
		if m.Approve {
			status = "handoff_ready"
		}
		if m.Installed {
			status = "installed"
		}
		send(w, fd.OnboardingStatus{Status: status, PollAfterSeconds: 1})
	case "/v2/agent-connections/acn_offline1/handoff":
		send(w, fd.Handoff{Status: "delivered", Credentials: &fd.AgentCredentials{ConnectionID: "acn_offline1", APIKey: "fd_candidate_offline", Generation: 2}})
	case "/v2/agent-connections/acn_offline1/identity":
		owner := strings.Contains(r.Header.Get("Authorization"), "candidate")
		if owner {
			send(w, map[string]any{"connection": fd.AgentConnection{ID: "acn_offline1", Ownership: "owned", Lifecycle: "valid", Access: "ready", Generation: 2, Installed: m.Installed, Scopes: sourceList}})
		} else {
			send(w, map[string]any{"connection": fd.AgentConnection{ID: "acn_offline1", Ownership: "unowned", Lifecycle: "valid", Access: "ready", Generation: 1, Scopes: sourceList}})
		}
	case "/v2/agent-connections/acn_offline1/handoff/ack":
		m.ACKCount++
		m.Installed = true
		if m.DropACK {
			m.DropACK = false
			conn, _, _ := w.(http.Hijacker).Hijack()
			_ = conn.Close()
			return
		}
		send(w, fd.Handoff{Status: "confirmed", Generation: 2})
	case "/v2/bulk-jobs":
		if r.Method == "GET" {
			send(w, map[string]any{"data": []any{}, "pagination": map[string]any{"hasMore": false}})
			return
		}
		m.CreateCount++
		var req fd.CreateBulkJobRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		m.BulkSource = string(req.Source)
		m.BulkStatus = "awaiting_upload"
		send(w, fd.BulkJobCreated{Job: fd.BulkJob{ID: "bulk_offline1", Source: m.BulkSource, Status: m.BulkStatus}, Upload: fd.BulkJobCreatedUpload{Href: m.URL + "/upload?signature=private", Method: "PUT", ContentType: "text/plain; charset=utf-8", ExpiresAt: time.Now().Add(time.Minute).UTC().Format(time.RFC3339)}})
	case "/upload":
		if key != "" {
			panic("upload leaked key")
		}
		m.UploadCount++
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(200)
	case "/v2/bulk-jobs/bulk_offline1":
		if r.Method == "DELETE" {
			m.BulkStatus = "cancelled"
		}
		send(w, map[string]any{"job": fd.BulkJob{ID: "bulk_offline1", Source: m.BulkSource, Status: m.BulkStatus}})
	case "/v2/bulk-jobs/bulk_offline1/start":
		m.StartCount++
		m.BulkStatus = "completed"
		if m.DropStart {
			m.DropStart = false
			conn, _, _ := w.(http.Hijacker).Hijack()
			_ = conn.Close()
			return
		}
		send(w, map[string]any{"job": fd.BulkJob{ID: "bulk_offline1", Source: m.BulkSource, Status: m.BulkStatus}})
	case "/v2/bulk-jobs/bulk_offline1/pause":
		m.BulkStatus = "paused"
		send(w, map[string]any{"job": fd.BulkJob{ID: "bulk_offline1", Source: m.BulkSource, Status: m.BulkStatus}})
	case "/v2/bulk-jobs/bulk_offline1/resume":
		m.BulkStatus = "completed"
		send(w, map[string]any{"job": fd.BulkJob{ID: "bulk_offline1", Source: m.BulkSource, Status: m.BulkStatus}})
	case "/v2/bulk-jobs/bulk_offline1/results":
		send(w, fd.BulkManifest{ID: "bulk_offline1", Parts: []fd.BulkManifestPart{{Number: 0}}})
	case "/v2/bulk-jobs/bulk_offline1/results/0":
		w.Header().Set("Location", m.URL+"/signed/part")
		w.WriteHeader(302)
	default:
		w.WriteHeader(404)
		send(w, map[string]string{"code": "not_found"})
	}
}
func invoke(t *testing.T, root, url, input string, args ...string) (int, string, string) {
	t.Helper()
	var out, stderr bytes.Buffer
	a := New(IO{In: strings.NewReader(input), Out: &out, Err: &stderr}, "1.0.0", "test")
	base := []string{"--state-dir", root, "--credential-store", "file", "--json", "--non-interactive"}
	if url != "" {
		base = append(base, "--data-origin", url, "--management-origin", url, "--web-origin", url)
	}
	code := a.Execute(context.Background(), append(base, args...))
	return code, out.String(), stderr.String()
}
func login(t *testing.T, root, url, key string) {
	t.Helper()
	code, _, err := invoke(t, root, url, key+"\n", "auth", "login", "--api-key-stdin")
	if code != 0 {
		t.Fatal(err)
	}
}
func stateStore(t *testing.T, root string) *state.Store {
	t.Helper()
	s, err := state.New(root, "default", "file", false)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestPublicPlansAndStatusNeedNoLogin(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "fd_must_not_reach_public_routes")
	mock, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	for _, item := range []struct{ command, expected string }{{"plans", `"currency":"USD"`}, {"status", `"status":"operational"`}} {
		code, out, stderr := invoke(t, root, server.URL, "", "discover", item.command)
		if code != 0 || !strings.Contains(out, item.expected) {
			t.Fatal(item.command, out, stderr)
		}
	}
	if mock.stats().AccessReads != 0 || mock.stats().IdentityReads != 0 {
		t.Fatal("public discovery unexpectedly checked authenticated access")
	}
}

func TestOfflineHelpVersionAndCommands(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	root := filepath.Join(t.TempDir(), "unused")
	for _, args := range [][]string{{"--version"}, {"--help"}, {"commands", "--json"}} {
		code, out, err := invoke(t, root, "", "", args...)
		if code != 0 || out == "" || err != "" {
			t.Fatal(code, out, err)
		}
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("offline commands created state")
	}
}
func TestQueryReplayConflictRevocationAndProvenance(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, server := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	input := `{"source":"google_search","query":"batteries","num":5}`
	code, out, err := invoke(t, root, server.URL, input, "query", "--input", "-", "--run-id", "research:2026-10-22", "--recurring")
	if code != 0 {
		t.Fatal(err)
	}
	if !strings.Contains(out, "9007199254740993") || !strings.Contains(out, "https://example.com/?q=provenance") {
		t.Fatal("lost source fields or large integer", out)
	}
	code, out, err = invoke(t, root, server.URL, `{"query":"batteries","source":"google_search","params":{"num":5}}`, "query", "--input", "-", "--run-id", "research:2026-10-22", "--recurring")
	if code != 0 || !strings.Contains(out, `"cached":true`) {
		t.Fatal(code, out, err)
	}
	if m.stats().QueryCount != 1 {
		t.Fatal("duplicate paid submission")
	}
	code, _, err = invoke(t, root, server.URL, `{"source":"google_search","query":"different"}`, "query", "--input", "-", "--run-id", "research:2026-10-22")
	if code == 0 || !strings.Contains(err, "run_id_input_conflict") {
		t.Fatal(err)
	}
	m.set(func(m *mockAPI) { m.Revoked = true })
	code, out, err = invoke(t, root, server.URL, input, "query", "--input", "-", "--run-id", "research:2026-10-22")
	if code == 0 || out != "" || !strings.Contains(err, "invalid_api_key") {
		t.Fatal(code, out, err)
	}
	if m.stats().QueryCount != 1 {
		t.Fatal("revoked execution submitted")
	}
}
func TestUnknownSubmissionIsDurableAndNotRepeated(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, s.URL, "fd_owner_offline")
	m.set(func(m *mockAPI) { m.DropQuery = true })
	for i := 0; i < 2; i++ {
		code, _, err := invoke(t, root, s.URL, `{"source":"google_search","query":"test"}`, "query", "--input", "-", "--run-id", "uncertain")
		if code == 0 {
			t.Fatal("unknown submission was successful")
		}
		if i == 0 && !strings.Contains(err, "request_outcome_unknown") {
			t.Fatal(err)
		}
	}
	if m.stats().QueryCount != 1 {
		t.Fatal("uncertain submission retried")
	}
	r, err := stateStore(t, root).ReadRun("uncertain")
	if err != nil || r.Status != "uncertain" {
		t.Fatal(r, err)
	}
}
func TestValidationAndDryRunDoNotSubmit(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, s.URL, "fd_owner_offline")
	for _, input := range []string{`{"source":"google_search","query":"x","unexpected":true}`, `{"source":"google_search","query":"x","num":5,"params":{"num":5}}`, `{"source":"google_search","query":"x","params":null}`, `{"source":"youtube_metadata","video_id":"bad"}`, `{"source":"google_search","query":"x","num":101}`} {
		code, _, err := invoke(t, root, s.URL, input, "query", "--input", "-")
		if code == 0 {
			t.Fatal("accepted invalid input", input, err)
		}
	}
	code, out, err := invoke(t, root, s.URL, `{"source":"google_search","query":"x","num":5}`, "query", "--input", "-", "--dry-run")
	if code != 0 || !strings.Contains(out, "estimate") {
		t.Fatal(code, out, err)
	}
	if m.stats().QueryCount != 0 {
		t.Fatal("dry-run submitted")
	}
}
func TestOnboardingProofAndInstallationBoundaries(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	code, out, err := invoke(t, root, s.URL, "", "init", "--name", "Research assistant", "--scopes", "google_search,youtube_metadata")
	if code != 0 {
		t.Fatal(out, err)
	}
	if strings.Contains(out, "fd_trial") || strings.Contains(out, "fda_setup") {
		t.Fatal("init leaked credentials")
	}
	code, _, err = invoke(t, root, s.URL, "", "init")
	if code != 0 || m.stats().RegisterCount != 1 {
		t.Fatal(err, "trial re-registered")
	}
	code, _, err = invoke(t, root, s.URL, "test", "search", "query", "--recurring", "--run-id", "scheduled")
	if code == 0 || !strings.Contains(err, "owned_access_required") {
		t.Fatal(err)
	}
	code, out, err = invoke(t, root, s.URL, "", "auth", "claim")
	if code != 0 || !strings.Contains(out, "confirmationCode") {
		t.Fatal(out, err)
	}
	m.set(func(m *mockAPI) { m.Approve = true })
	var stdout, stderr bytes.Buffer
	a := New(IO{In: strings.NewReader(""), Out: &stdout, Err: &stderr}, "1.0.0", "test")
	a.VerifyFresh = func(ctx context.Context, store *state.Store, v *state.Vault) error {
		saved, e := store.Load(ctx)
		if e != nil {
			return e
		}
		if saved.Phase != "installing" || saved.Credentials.APIKey != "fd_candidate_offline" {
			t.Error("candidate not saved before verification")
		}
		if m.stats().ACKCount != 0 {
			t.Error("ACK before verification")
		}
		return nil
	}
	args := []string{"--state-dir", root, "--credential-store", "file", "--json", "--non-interactive", "auth", "finish", "--wait"}
	if code = a.Execute(context.Background(), args); code != 0 {
		t.Fatal(stderr.String())
	}
	if m.stats().ACKCount != 1 {
		t.Fatal("missing ACK")
	}
	saved, e := stateStore(t, root).Load(context.Background())
	if e != nil || saved.Phase != "installed" || saved.Credentials.HandoffProof != "" || saved.RegistrationProof != "" {
		t.Fatal("obsolete secrets remain", e)
	}
	code, _, err = invoke(t, root, s.URL, "", "auth", "logout")
	if code != 0 {
		t.Fatal(err)
	}
	code, _, err = invoke(t, root, s.URL, "", "init", "--name", "Replacement")
	if code == 0 || m.stats().RegisterCount != 1 {
		t.Fatal("logout created another trial", err)
	}
}
func TestAsyncAndBulkRecovery(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, s.URL, "fd_owner_offline")
	code, out, err := invoke(t, root, s.URL, `{"source":"youtube_metadata","video_id":"dQw4w9WgXcQ"}`, "query", "--input", "-", "--mode", "async", "--run-id", "async-1")
	if code != 0 || !strings.Contains(out, "job_offline") {
		t.Fatal(out, err)
	}
	code, out, err = invoke(t, root, s.URL, "", "runs", "recover", "async-1")
	if code != 0 || !strings.Contains(out, "9007199254740993") {
		t.Fatal(out, err)
	}
	ids := filepath.Join(t.TempDir(), "ids.txt")
	_ = os.WriteFile(ids, []byte("dQw4w9WgXcQ\n"), 0600)
	outFile := filepath.Join(t.TempDir(), "receipts.jsonl")
	code, out, err = invoke(t, root, s.URL, "", "bulk", "run", "--input", ids, "--run-id", "bulk-1", "--output", outFile)
	if code != 0 {
		t.Fatal(out, err)
	}
	if m.stats().CreateCount != 1 || m.stats().UploadCount != 1 || m.stats().StartCount != 1 {
		t.Fatal("incomplete bulk flow")
	}
	b, _ := os.ReadFile(outFile)
	if !strings.Contains(string(b), "9007199254740993123") {
		t.Fatal("bulk integer lost")
	}
	code, _, err = invoke(t, root, s.URL, "", "bulk", "run", "--input", ids, "--run-id", "bulk-1")
	if code != 0 || m.stats().CreateCount != 1 || m.stats().UploadCount != 1 {
		t.Fatal("bulk replay submitted", err)
	}
}
func TestBulkPartialReceiptsAndExistingResultsRevalidate(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	m.set(func(m *mockAPI) { m.Partial = true })
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, s.URL, "fd_owner_offline")
	ids := filepath.Join(t.TempDir(), "ids.txt")
	_ = os.WriteFile(ids, []byte("dQw4w9WgXcQ\n"), 0600)
	outFile := filepath.Join(t.TempDir(), "partial.jsonl")
	code, _, err := invoke(t, root, s.URL, "", "bulk", "run", "--input", ids, "--run-id", "partial", "--output", outFile)
	if code == 0 || !strings.Contains(err, "bulk_item_failures") {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(outFile)
	if !strings.Contains(string(b), "invalid") {
		t.Fatal("failed receipts lost")
	}
	m.set(func(m *mockAPI) { m.Revoked = true })
	code, out, err := invoke(t, root, s.URL, "", "runs", "show", "partial", "--result")
	if code == 0 || out != "" {
		t.Fatal("revoked cached result read", err)
	}
}
func TestIdentityUnavailableAndKeyIsolation(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	m, s := fixture(t)
	m.set(func(m *mockAPI) { m.IdentityMissing = true })
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, s.URL, "fd_owner_offline")
	code, _, err := invoke(t, root, s.URL, "", "search", "once", "--run-id", "oneoff")
	if code != 0 {
		t.Fatal(err)
	}
	code, _, err = invoke(t, root, s.URL, "", "search", "daily", "--recurring", "--run-id", "daily")
	if code == 0 || !strings.Contains(err, "identity_unavailable") {
		t.Fatal(err)
	}
	code, out, err := invoke(t, root, s.URL, "", "doctor")
	if code != 0 || !strings.Contains(out, "key_access_only") {
		t.Fatal(out, err)
	}
	code, _, err = invoke(t, root, s.URL, "", "auth", "login", "fd_owner_offline")
	if code == 0 || strings.Contains(err, "fd_owner_offline") {
		t.Fatal("key accepted/echoed as argument", err)
	}
}
