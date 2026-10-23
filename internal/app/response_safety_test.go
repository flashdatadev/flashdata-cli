package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
)

func TestMarkdownCredentialRedactionInSavedAndExportedResponses(t *testing.T) {
	fixture := readResponses(t).Redaction
	for _, example := range fixture.Responses {
		for _, mode := range []string{"stdout", "json", "file"} {
			t.Run(example.Accept+"/"+mode, func(t *testing.T) {
				var out bytes.Buffer
				a := New(IO{Out: &out}, "test", "test")
				a.JSON = mode == "json"
				a.knownSecrets = fixture.Secrets
				v := &fd.FormattedResponse{JSON: example.Response}
				if example.Accept == "text/plain" {
					v.JSON = nil
					if err := json.Unmarshal(example.Response, &v.Text); err != nil {
						t.Fatal(err)
					}
				}
				s := stateStore(t, filepath.Join(t.TempDir(), "state"))
				if err := s.Ensure(); err != nil {
					t.Fatal(err)
				}
				r := &state.Run{ID: "redacted-result"}
				body, err := a.saveResponse(s, r, v, responseFlags{Accept: example.Accept, RespondWith: "markdown"})
				if err != nil {
					t.Fatal(err)
				}
				saved, err := s.ReadResult(r)
				if err != nil || !bytes.Equal(body, saved) {
					t.Fatal("redacted result was not saved", err)
				}
				output := ""
				if mode == "file" {
					output = filepath.Join(t.TempDir(), "result.md")
				}
				if err := a.emitResult(r, body, output, false); err != nil {
					t.Fatal(err)
				}
				if output != "" {
					body, err = os.ReadFile(output)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					body = out.Bytes()
				}
				for _, b := range [][]byte{body, saved} {
					if bytes.Contains(bytes.ToLower(b), []byte("credentialdummy")) {
						t.Fatal("credential survived Markdown escaping", string(b))
					}
					if !bytes.Contains(b, []byte("https://example.com/article?q=provenance")) {
						t.Fatal("provenance link was changed", string(b))
					}
				}
				text := string(saved)
				if example.Accept == "application/json" {
					var document struct{ Content string }
					if err := json.Unmarshal(saved, &document); err != nil {
						t.Fatal(err)
					}
					text = document.Content
				}
				if !strings.Contains(text, `Unrelated \*\_\[\] &amp; &lt;&gt;`) {
					t.Fatal("unrelated Markdown was rewritten", text)
				}
			})
		}
	}
}

func TestOlderCachedMarkdownIsRedactedOnExport(t *testing.T) {
	fixture := readResponses(t).Redaction
	for _, example := range fixture.Responses {
		for _, batch := range []bool{false, true} {
			if batch && example.Accept == "text/plain" {
				continue
			}
			for _, mode := range []string{"stdout", "json", "file"} {
				t.Run(fmt.Sprintf("%s/batch=%v/%s", example.Accept, batch, mode), func(t *testing.T) {
					body := example.Response
					format := ""
					if batch {
						body, _ = json.Marshal([]json.RawMessage{body, body})
					}
					if example.Accept == "text/plain" {
						var text string
						if err := json.Unmarshal(body, &text); err != nil {
							t.Fatal(err)
						}
						body, format = []byte(text), "text"
					}
					s := stateStore(t, filepath.Join(t.TempDir(), "state"))
					if err := s.Ensure(); err != nil {
						t.Fatal(err)
					}
					r := &state.Run{ID: "older-markdown", ResultFormat: format, Status: "succeeded", ResultHash: state.Digest(body)}
					if err := state.AtomicWrite(s.ResultPath(r.ID), body); err != nil {
						t.Fatal(err)
					}
					if err := s.SaveRun(r); err != nil {
						t.Fatal(err)
					}
					var out bytes.Buffer
					a := New(IO{Out: &out}, "test", "test")
					a.JSON, a.knownSecrets = mode == "json", fixture.Secrets
					output := ""
					if mode == "file" {
						output = filepath.Join(t.TempDir(), "cached.md")
					}
					if err := a.exportRun(s, r, output, true); err != nil {
						t.Fatal(err)
					}
					body = out.Bytes()
					if output != "" {
						var err error
						body, err = os.ReadFile(output)
						if err != nil {
							t.Fatal(err)
						}
					}
					if bytes.Contains(bytes.ToLower(body), []byte("credentialdummy")) {
						t.Fatal("older cached credential was exported", string(body))
					}
				})
			}
		}
	}
}

type responseTestTransport func(*http.Request) (*http.Response, error)

func (f responseTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFormatChangeJournalFailurePreservesCachedResponse(t *testing.T) {
	t.Setenv("FLASHDATA_API_KEY", "")
	for _, batch := range []bool{false, true} {
		t.Run(map[bool]string{false: "async", true: "batch"}[batch], func(t *testing.T) {
			mock, server, _ := formatFixture(t)
			root := filepath.Join(t.TempDir(), "state")
			login(t, root, server.URL, "fd_owner_offline")
			input := ""
			args := []string{"search", "offline", "--mode", "async", "--wait", "--run-id", "safe-reformat"}
			if batch {
				b, _ := json.Marshal(readResponses(t).Batch.Queries)
				input = string(b)
				args = []string{"batch", "submit", "--input", "-", "--wait", "--run-id", "safe-reformat"}
			}
			code, original, stderr := invokeFormat(t, root, server.URL, input, true, args...)
			if code != 0 {
				t.Fatal(stderr)
			}
			s := stateStore(t, root)
			path := s.RunPath("safe-reformat")
			backup := path + ".saved"
			faulted := false
			restore := func() {
				if faulted {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					if err := os.Rename(backup, path); err != nil {
						t.Fatal(err)
					}
					faulted = false
				}
			}
			t.Cleanup(restore)
			var out, diagnostic bytes.Buffer
			a := New(IO{In: strings.NewReader(""), Out: &out, Err: &diagnostic}, "test", "test")
			a.HTTPClient = &http.Client{Transport: responseTestTransport(func(request *http.Request) (*http.Response, error) {
				response, err := http.DefaultTransport.RoundTrip(request)
				if err == nil && strings.HasSuffix(request.URL.Path, "/results") && !faulted {
					if err = os.Rename(path, backup); err == nil {
						err = os.Mkdir(path, 0700)
						if err != nil {
							_ = os.Rename(backup, path)
						} else {
							faulted = true
						}
					}
					if err != nil {
						response.Body.Close()
						return nil, err
					}
				}
				return response, err
			})}
			code = a.Execute(context.Background(), []string{"--state-dir", root, "--credential-store", "file", "--non-interactive", "--json", "runs", "recover", "safe-reformat", "--accept", "text/plain"})
			if code == 0 || !faulted {
				t.Fatal("journal failure was not exercised", diagnostic.String())
			}
			restore()
			before := mock.stats().QueryCount
			for _, command := range [][]string{{"runs", "recover", "safe-reformat"}, {"runs", "show", "safe-reformat", "--result"}} {
				code, result, stderr := invokeFormat(t, root, server.URL, "", true, command...)
				if code != 0 {
					t.Fatal("previous cache was damaged", stderr)
				}
				var want, got Envelope
				if json.Unmarshal([]byte(original), &want) != nil || json.Unmarshal([]byte(result), &got) != nil {
					t.Fatal("invalid cached envelope")
				}
				wantBytes, _ := json.Marshal(want.Data)
				gotBytes, _ := json.Marshal(got.Data)
				if !bytes.Equal(wantBytes, gotBytes) || !got.Cached {
					t.Fatal("saved response changed after journal failure", result)
				}
			}
			code, _, stderr = invokeFormat(t, root, server.URL, "", true, "runs", "recover", "safe-reformat", "--accept", "text/plain")
			if code != 0 || mock.stats().QueryCount != before {
				t.Fatal("safe format retry failed or resubmitted work", stderr)
			}
			r, err := s.ReadRun("safe-reformat")
			if err != nil || r.ResponseAccept != "text/plain" || r.ResultFormat != "text" {
				t.Fatal("new format was not committed", err)
			}
			f, err := s.OpenResult(r)
			if err != nil {
				t.Fatal(err)
			}
			b, err := io.ReadAll(f)
			f.Close()
			if err != nil || !bytes.Contains(b, []byte("Source:")) {
				t.Fatal("new result version is unreadable", err)
			}
		})
	}
}
