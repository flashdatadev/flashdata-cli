package app

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	fd "github.com/flashdatadev/flashdata-go"
)

type responseCase struct {
	Accept, RespondWith string
	Response            json.RawMessage
}
type responseSource struct {
	Query                                fd.Query
	Realtime, Stored, Submission, Status []responseCase
}
type responseFixtures struct {
	Sources map[fd.Source]responseSource
	Batch   struct {
		Queries   []fd.Query
		Responses []responseCase
	}
}

//go:embed testdata/response-formats.json
var responseFixtureJSON []byte

func readResponses(t *testing.T) responseFixtures {
	t.Helper()
	var result responseFixtures
	if err := json.Unmarshal(responseFixtureJSON, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

type wireCall struct {
	Path, Accept, RespondWith string
	Body                      []byte
}
type responseAPI struct {
	mu            sync.Mutex
	Calls         []wireCall
	ProblemStatus int
	TextOverride  *string
}

func (a *responseAPI) calls() []wireCall {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]wireCall(nil), a.Calls...)
}
func (a *responseAPI) configure(status int, text *string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ProblemStatus = status
	a.TextOverride = text
}

func formatFixture(t *testing.T) (*mockAPI, *httptest.Server, *responseAPI) {
	t.Helper()
	fixtures := readResponses(t)
	mock, _ := fixture(t)
	wire := &responseAPI{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		wire.mu.Lock()
		wire.Calls = append(wire.Calls, wireCall{r.URL.Path, r.Header.Get("Accept"), r.Header.Get("X-Respond-With"), body})
		problem, override := wire.ProblemStatus, wire.TextOverride
		wire.mu.Unlock()
		if problem > 0 && (strings.HasPrefix(r.URL.Path, "/v1/queries") || strings.HasSuffix(r.URL.Path, "/results")) {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(problem)
			_, _ = w.Write([]byte(`{"code":"format_rejected","detail":"test problem"}`))
			return
		}
		recorder := httptest.NewRecorder()
		mock.handler(recorder, r)
		if recorder.Code < 200 || recorder.Code >= 400 {
			for k, v := range recorder.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(recorder.Code)
			_, _ = w.Write(recorder.Body.Bytes())
			return
		}
		accept, representation := r.Header.Get("Accept"), r.Header.Get("X-Respond-With")
		var cases []responseCase
		id := "job_offline"
		var q fd.Query
		_ = json.Unmarshal(body, &q)
		switch {
		case r.URL.Path == "/v1/queries/realtime":
			cases = fixtures.Sources[q.Source].Realtime
		case r.URL.Path == "/v1/queries":
			cases = fixtures.Sources[q.Source].Submission
		case r.URL.Path == "/v1/queries/batch":
			cases = fixtures.Batch.Responses
		case strings.HasPrefix(r.URL.Path, "/v2/jobs/") || strings.HasPrefix(r.URL.Path, "/signed/job_"):
			id = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/v2/jobs/"), "/signed/"), "/results")
			mock.mu.Lock()
			source := mock.Jobs[id]
			mock.mu.Unlock()
			if source == "" {
				source = fd.SourceYouTubeMetadata
			}
			if strings.HasSuffix(r.URL.Path, "/results") && accept == "application/json" && representation == "" {
				for k, v := range recorder.Header() {
					w.Header()[k] = v
				}
				w.WriteHeader(recorder.Code)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/signed/") || strings.HasSuffix(r.URL.Path, "/results") {
				cases = fixtures.Sources[source].Stored
			} else {
				cases = fixtures.Sources[source].Status
			}
		}
		if cases == nil {
			for k, v := range recorder.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(recorder.Code)
			_, _ = w.Write(recorder.Body.Bytes())
			return
		}
		for _, example := range cases {
			if example.Accept == accept && example.RespondWith == representation {
				w.Header().Set("Content-Type", accept+"; charset=utf-8")
				response := strings.ReplaceAll(string(example.Response), "job_offline", id)
				if accept == "text/plain" {
					var text string
					_ = json.Unmarshal([]byte(response), &text)
					if override != nil {
						text = *override
					}
					_, _ = io.WriteString(w, text)
				} else {
					_, _ = io.WriteString(w, response)
				}
				return
			}
		}
		t.Error("missing API format fixture", r.URL.Path, accept, representation)
		w.WriteHeader(500)
	}))
	mock.set(func(m *mockAPI) { m.URL = server.URL })
	t.Cleanup(server.Close)
	return mock, server, wire
}

func invokeFormat(t *testing.T, root, url, input string, envelope bool, args ...string) (int, string, string) {
	t.Helper()
	var out, stderr bytes.Buffer
	a := New(IO{In: strings.NewReader(input), Out: &out, Err: &stderr}, "1.0.1", "test")
	base := []string{"--state-dir", root, "--credential-store", "file", "--non-interactive"}
	if envelope {
		base = append(base, "--json")
	}
	if url != "" {
		base = append(base, "--data-origin", url, "--management-origin", url, "--web-origin", url)
	}
	code := a.Execute(context.Background(), append(base, args...))
	return code, out.String(), stderr.String()
}
func expectedResponse(t *testing.T, out string, example responseCase, envelope bool) {
	t.Helper()
	want := example.Response
	got := []byte(out)
	if envelope || example.Accept == "application/json" {
		var response struct {
			OK   bool
			Data json.RawMessage
		}
		if err := json.Unmarshal(got, &response); err != nil || !response.OK {
			t.Fatal("invalid CLI envelope", out)
		}
		got = response.Data
	}
	if example.Accept == "text/plain" && !envelope {
		var text string
		_ = json.Unmarshal(want, &text)
		if string(got) != text {
			t.Fatal("API text changed", string(got), text)
		}
		return
	}
	var a, b any
	_ = json.Unmarshal(got, &a)
	_ = json.Unmarshal(want, &b)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("API response fields changed", string(got))
	}
}

func TestResponseFormatsAllRealtimeSources(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	fixtures := readResponses(t)
	mock, server, wire := formatFixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	for source, data := range fixtures.Sources {
		for i, example := range data.Realtime {
			t.Run(fmt.Sprintf("%s/%s/%s", source, example.Accept, example.RespondWith), func(t *testing.T) {
				input, _ := json.Marshal(data.Query)
				args := []string{"query", "--input", "-", "--run-id", fmt.Sprintf("%s-%d", source, i), "--accept", example.Accept}
				if example.RespondWith != "" {
					args = append(args, "--respond-with", example.RespondWith)
				}
				code, out, stderr := invokeFormat(t, root, server.URL, string(input), false, args...)
				if code != 0 {
					t.Fatal(stderr)
				}
				expectedResponse(t, out, example, false)
				before := mock.stats().QueryCount
				code, cached, stderr := invokeFormat(t, root, server.URL, string(input), false, args...)
				if code != 0 || mock.stats().QueryCount != before {
					t.Fatal("cache changed or repeated query", stderr)
				}
				expectedResponse(t, cached, example, false)
			})
		}
	}
	for _, call := range wire.calls() {
		if !strings.HasPrefix(call.Path, "/v1/queries") && call.RespondWith != "" {
			t.Fatal("representation leaked into preflight")
		}
	}
}

func TestResponseFormatsAllStoredSources(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	fixtures := readResponses(t)
	mock, server, _ := formatFixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	for source, data := range fixtures.Sources {
		mock.set(func(m *mockAPI) { m.Jobs = map[string]fd.Source{"job_offline": source} })
		for _, example := range data.Stored {
			t.Run(fmt.Sprintf("%s/%s/%s", source, example.Accept, example.RespondWith), func(t *testing.T) {
				args := []string{"jobs", "result", "job_offline", "--accept", example.Accept}
				if example.RespondWith != "" {
					args = append(args, "--respond-with", example.RespondWith)
				}
				code, out, stderr := invokeFormat(t, root, server.URL, "", false, args...)
				if code != 0 {
					t.Fatal(stderr)
				}
				expectedResponse(t, out, example, false)
			})
		}
	}
	if mock.stats().QueryCount != 0 {
		t.Fatal("result read submitted work")
	}
}

func TestTextReceiptRecoveryAndBatch(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	fixtures := readResponses(t)
	mock, server, wire := formatFixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	for source, data := range fixtures.Sources {
		input, _ := json.Marshal(data.Query)
		run := "receipt-" + string(source)
		code, out, stderr := invokeFormat(t, root, server.URL, string(input), false, "query", "--input", "-", "--mode", "async", "--accept", "text/plain", "--run-id", run)
		if code != 0 {
			t.Fatal(stderr)
		}
		expectedResponse(t, out, data.Submission[1], false)
		r, err := stateStore(t, root).ReadRun(run)
		if err != nil || r.RemoteID != "job_offline" || r.Status != "accepted" || r.ReceiptAccept != "text/plain" {
			t.Fatal("text receipt lost durable acceptance", r, err)
		}
		code, _, stderr = invokeFormat(t, root, server.URL, "", true, "runs", "recover", run, "--accept", "application/json", "--respond-with", "markdown")
		if code != 0 {
			t.Fatal(stderr)
		}
	}
	input, _ := json.Marshal(fixtures.Batch.Queries)
	code, out, stderr := invokeFormat(t, root, server.URL, string(input), false, "batch", "submit", "--input", "-", "--accept", "text/plain", "--run-id", "text-batch")
	if code != 0 {
		t.Fatal(stderr)
	}
	expectedResponse(t, out, fixtures.Batch.Responses[1], false)
	r, err := stateStore(t, root).ReadRun("text-batch")
	if err != nil || !reflect.DeepEqual(r.RemoteIDs, []string{"job_offline", "job_offline1"}) {
		t.Fatal("batch IDs lost", r, err)
	}
	before := mock.stats().QueryCount
	code, out, stderr = invokeFormat(t, root, server.URL, "", false, "runs", "recover", "text-batch", "--respond-with", "markdown")
	if code != 0 || !strings.Contains(out, "Job: job_offline1") || !strings.Contains(out, "Markdown Content:") || mock.stats().QueryCount != before {
		t.Fatal("batch result formatting failed", out, stderr)
	}
	for _, call := range wire.calls() {
		if call.Path == "/v1/queries" || call.Path == "/v1/queries/batch" {
			if call.Accept != "text/plain" || call.RespondWith != "" {
				t.Fatal("receipt header scope", call)
			}
		}
	}
}

func TestTextWithJSONEnvelopeStatusOutputAndCacheConflict(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	fixtures := readResponses(t)
	mock, server, _ := formatFixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	mock.set(func(m *mockAPI) { m.Jobs = map[string]fd.Source{"job_offline": fd.SourceGoogleSearch} })
	for _, command := range []string{"get", "wait"} {
		code, out, stderr := invokeFormat(t, root, server.URL, "", true, "jobs", command, "job_offline", "--accept", "text/plain")
		if code != 0 {
			t.Fatal(stderr)
		}
		expectedResponse(t, out, fixtures.Sources[fd.SourceGoogleSearch].Status[1], true)
	}
	path := filepath.Join(t.TempDir(), "result.md")
	code, _, stderr := invokeFormat(t, root, server.URL, "", false, "search", "offline", "--accept", "text/plain", "--respond-with", "markdown", "--output", path, "--run-id", "cached-md")
	if code != 0 {
		t.Fatal(stderr)
	}
	b, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(b), "Markdown Content:") {
		t.Fatal("wrong output file", err)
	}
	code, out, stderr := invokeFormat(t, root, server.URL, "", true, "search", "offline", "--run-id", "cached-md")
	if code != 0 {
		t.Fatal(stderr)
	}
	expectedResponse(t, out, fixtures.Sources[fd.SourceGoogleSearch].Realtime[3], true)
	before := mock.stats().QueryCount
	code, out, stderr = invokeFormat(t, root, server.URL, "", true, "search", "offline", "--run-id", "cached-md", "--accept", "application/json")
	if code != 1 || out != "" || !strings.Contains(stderr, "run_response_format_conflict") || mock.stats().QueryCount != before {
		t.Fatal("cached format change repeated work", stderr)
	}
}

func TestResponseErrorsInvalidFlagsAndMalformedReceipt(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	mock, server, wire := formatFixture(t)
	root := filepath.Join(t.TempDir(), "state")
	login(t, root, server.URL, "fd_owner_offline")
	before := len(wire.calls())
	for _, args := range [][]string{
		{"search", "offline", "--accept", "text/html"},
		{"search", "offline", "--respond-with", "json"},
		{"search", "offline", "--mode", "async", "--respond-with", "markdown"},
		{"batch", "submit", "--input", "-", "--respond-with", "markdown"},
		{"jobs", "get", "job_offline", "--respond-with", "markdown"},
		{"account", "balance", "--accept", "text/plain"},
	} {
		code, out, stderr := invokeFormat(t, root, server.URL, "[]", true, args...)
		if code != 1 || out != "" || stderr == "" {
			t.Fatal("invalid flags accepted", args, stderr)
		}
	}
	if len(wire.calls()) != before {
		t.Fatal("invalid formats performed I/O")
	}
	for _, status := range []int{400, 406, 429} {
		wire.configure(status, nil)
		code, out, stderr := invokeFormat(t, root, server.URL, "", false, "search", "offline", "--accept", "text/plain", "--run-id", fmt.Sprintf("problem-%d", status))
		if code != 1 || out != "" || !strings.Contains(stderr, "format_rejected") {
			t.Fatal("lost API error", stderr)
		}
	}
	bad := `Id: "job_candidate"` + "\nSource: \"wrong\"\nStatus: \"pending\"\n"
	wire.configure(0, &bad)
	code, out, stderr := invokeFormat(t, root, server.URL, "", false, "search", "offline", "--mode", "async", "--accept", "text/plain", "--run-id", "bad-receipt")
	if code != 1 || out != "" || !strings.Contains(stderr, "job_candidate") {
		t.Fatal("malformed receipt confirmed", stderr)
	}
	r, err := stateStore(t, root).ReadRun("bad-receipt")
	if err != nil || r.Status != "uncertain" || r.RemoteID != "" || !reflect.DeepEqual(r.CandidateRemoteIDs, []string{"job_candidate"}) {
		t.Fatal("recovery evidence lost", r, err)
	}
	beforeQueries := mock.stats().QueryCount
	code, _, _ = invokeFormat(t, root, server.URL, "", false, "search", "offline", "--mode", "async", "--accept", "text/plain", "--run-id", "bad-receipt")
	if code != 1 || mock.stats().QueryCount != beforeQueries {
		t.Fatal("uncertain submission repeated")
	}
}

func TestTextReceiptBoundariesAndRedaction(t *testing.T) {
	q := []fd.Query{{Source: fd.SourceGoogleSearch}}
	for _, text := range []string{
		"https://example.com/job_looks_like_id", "Id: job_unquoted\nSource: \"google_search\"\nStatus: \"pending\"\n",
		"Id: \"job_duplicate\"\nId: \"job_duplicate\"\nSource: \"google_search\"\nStatus: \"pending\"\n",
		"Id: \"job_valid\"\nSource: \"google_search\"\nStatus: null\n",
	} {
		_, err := textReceipts(text, q, false)
		var unknown *fd.SubmissionUnknownError
		if !errors.As(err, &unknown) {
			t.Fatal("malformed receipt accepted")
		}
	}
	a := New(IO{}, "test", "test")
	a.knownSecrets = []string{"my-private-key"}
	text := a.redactedResponseText("Secret key: \"unknown-secret\"\n**Access key:** another-secret\nmy-private-key\n<https://storage.example/file?X-Amz-Signature=private>\nhttps://example.com/?q=provenance\n")
	for _, secret := range []string{"unknown-secret", "another-secret", "my-private-key", "X-Amz-Signature"} {
		if strings.Contains(text, secret) {
			t.Fatal("text leaked credential", secret)
		}
	}
	if !strings.Contains(text, "https://example.com/?q=provenance") {
		t.Fatal("provenance URL lost")
	}
	raw, _ := json.Marshal(map[string]any{"source": "youtube_download", "status": "completed", "content": "# Download\n<https://storage.example/file?signature=private>", "usage": map[string]any{"credits": 0.01, "secret_key": "private-usage"}})
	body, _, err := a.responseBytes(&fd.FormattedResponse{JSON: raw})
	if err != nil || strings.Contains(string(body), "signature=private") {
		t.Fatal("Markdown JSON leaked signed URL", err)
	}
	usage := a.responseUsage(&fd.FormattedResponse{JSON: raw})
	if strings.Contains(string(usage), "private-usage") {
		t.Fatal("usage metadata leaked a secret")
	}
	_, err = textReceipts("Id: \"job_first\"\nId: \"job_second\"\nSource: \"google_search\"\nStatus: \"pending\"\n", q, false)
	var unknown *fd.SubmissionUnknownError
	if !errors.As(err, &unknown) || !reflect.DeepEqual(unknown.JobIDs, []string{"job_first", "job_second"}) {
		t.Fatal("duplicate IDs lost recovery evidence", err)
	}
}

func TestStandaloneResponseFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("compiled process acceptance")
	}
	t.Setenv("FLASHDATA_API_KEY", "")
	fixtures := readResponses(t)
	mock, server, _ := formatFixture(t)
	binary := buildBinary(t)
	root := filepath.Join(t.TempDir(), "state")
	code, _, stderr := binaryCall(t, binary, root, server.URL, "fd_owner_offline\n", "auth", "login", "--api-key-stdin")
	if code != 0 {
		t.Fatal(stderr)
	}
	for i, example := range fixtures.Sources[fd.SourceGoogleSearch].Realtime {
		args := []string{"search", "offline", "--run-id", fmt.Sprintf("binary-%d", i), "--accept", example.Accept}
		if example.RespondWith != "" {
			args = append(args, "--respond-with", example.RespondWith)
		}
		code, out, stderr := binaryCall(t, binary, root, server.URL, "", args...)
		if code != 0 {
			t.Fatal(stderr)
		}
		expectedResponse(t, out, example, true)
	}
	code, out, stderr := binaryCall(t, binary, root, server.URL, "", "youtube", "transcript", "dQw4w9WgXcQ", "--accept", "text/plain", "--respond-with", "markdown", "--json=false", "--run-id", "binary-raw-transcript")
	if code != 0 {
		t.Fatal(stderr)
	}
	expectedResponse(t, out, fixtures.Sources[fd.SourceYouTubeTranscript].Realtime[3], false)
	code, _, stderr = binaryCall(t, binary, root, server.URL, "", "search", "offline", "--mode", "async", "--wait", "--accept", "text/plain", "--respond-with", "markdown", "--run-id", "binary-wait")
	if code != 0 {
		t.Fatal(stderr)
	}
	before := mock.stats().QueryCount
	code, _, stderr = binaryCall(t, binary, root, server.URL, "", "runs", "recover", "binary-wait", "--accept", "application/json", "--respond-with", "")
	if code != 0 || mock.stats().QueryCount != before {
		t.Fatal("format recovery replayed query", stderr)
	}
}
