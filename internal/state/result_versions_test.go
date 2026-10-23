package state

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestResultVersionsKeepCommittedAndLegacyResponses(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(map[bool]string{false: "versioned", true: "legacy"}[legacy], func(t *testing.T) {
			s := newStore(t, "file")
			original := []byte(`{"source":"google_search","results":[]}`)
			r := &Run{ID: "safe-reformat", Kind: "query:async", ResponseAccept: "application/json"}
			if legacy {
				r.ResultHash, r.Status = Digest(original), "succeeded"
				if err := AtomicWrite(s.ResultPath(r.ID), original); err != nil {
					t.Fatal(err)
				}
				if err := s.SaveRun(r); err != nil {
					t.Fatal(err)
				}
			} else if err := s.SaveResult(r, original); err != nil {
				t.Fatal(err)
			}
			old := *r
			// A staged file and the old journal model an interruption before
			// the journal replacement, without relying on filesystem timing.
			replacement := []byte("Source: \"google_search\"\nResults: []\n")
			staged := filepath.Join(s.resultVersionDir(r.ID), Digest(replacement)+".json")
			if err := AtomicWrite(staged, replacement); err != nil {
				t.Fatal(err)
			}
			reloaded, err := s.ReadRun(r.ID)
			if err != nil {
				t.Fatal(err)
			}
			check := func(record *Run, expected []byte) {
				t.Helper()
				b, err := s.ReadResult(record)
				if err != nil || !bytes.Equal(b, expected) {
					t.Fatal("committed response changed", string(b), err)
				}
				file, err := s.OpenResult(record)
				if err != nil {
					t.Fatal(err)
				}
				b, err = io.ReadAll(file)
				file.Close()
				if err != nil || !bytes.Equal(b, expected) {
					t.Fatal("stream reader selected another version", err)
				}
			}
			check(reloaded, original)
			r.ResultFormat, r.ResponseAccept = "text", "text/plain"
			if err := s.SaveResult(r, replacement); err != nil {
				t.Fatal(err)
			}
			check(r, replacement)
			check(&old, original)
			if err := AtomicWrite(staged, []byte("tampered")); err != nil {
				t.Fatal(err)
			}
			if _, err := s.ReadResult(r); !errors.Is(err, ErrCorrupt) {
				t.Fatal("corrupt version accepted", err)
			}
		})
	}
}

func TestOrphanedResponseVersionBlocksNewIntent(t *testing.T) {
	s := newStore(t, "file")
	id := "staged-without-journal"
	if err := AtomicWrite(filepath.Join(s.resultVersionDir(id), Digest([]byte("result"))+".json"), []byte("result")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReadRun(id); !errors.Is(err, ErrCorrupt) {
		t.Fatal("orphaned version allowed another execution", err)
	}
	r := &Run{ID: id, ResultHash: "../../outside"}
	if _, err := s.ReadResult(r); !errors.Is(err, ErrCorrupt) {
		t.Fatal("unsafe result hash accepted", err)
	}
	if _, err := os.Stat(s.RunPath(id)); !os.IsNotExist(err) {
		t.Fatal("orphan inspection created a run", err)
	}
}
