// Package state owns private local storage. The SDK never handles local state.
package state

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"time"

	fd "github.com/flashdatadev/flashdata-go"
)

var ErrUnsafe = errors.New("unsafe_state_storage")
var ErrCorrupt = errors.New("state_corrupt_or_unsupported; preserve files and recover explicitly")
var ErrBusy = errors.New("profile_busy; retry after the other process exits")
var ErrVault = errors.New("credential_store_unavailable; unlock the OS store or explicitly select file mode for a new profile")
var ErrMissing = errors.New("credentials_missing; import a key with auth login")
var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,63}$`)
var validRun = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:@+-]{0,159}$`)
var validResultHash = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Store struct {
	Root, Dir, Profile, Mode string
	Interactive              bool
}
type Vault struct {
	originalCredentials *fd.AgentCredentials
	originalPhase       string
	Version             int                           `json:"version"`
	Phase               string                        `json:"phase"`
	DataOrigin          string                        `json:"dataOrigin"`
	ManagementOrigin    string                        `json:"managementOrigin"`
	WebOrigin           string                        `json:"webOrigin"`
	Credentials         fd.AgentCredentials           `json:"credentials"`
	RegistrationProof   string                        `json:"registrationProof,omitempty"`
	Registration        *fd.RegistrationRequest       `json:"registration,omitempty"`
	Claim               *fd.AgentClaim                `json:"claim,omitempty"`
	Uploads             map[string]*fd.BulkJobCreated `json:"uploads,omitempty"`
	BulkRuns            map[string]string             `json:"bulkRuns,omitempty"`
	MigratedFrom        string                        `json:"migratedFrom,omitempty"`
}

// UseEnvironmentKey changes only this command's access. Save restores the
// profile's original credentials while committing unrelated recovery updates.
func (v *Vault) UseEnvironmentKey(key string) {
	if v.originalCredentials == nil {
		original := v.Credentials
		v.originalCredentials = &original
		v.originalPhase = v.Phase
	}
	v.Credentials = fd.AgentCredentials{APIKey: key}
	v.Phase = "environment"
}

func (v Vault) String() string   { return "flashdata.Vault{redacted}" }
func (v Vault) GoString() string { return v.String() }

type envelope struct {
	Version    int    `json:"version"`
	Protection string `json:"protection"`
	Nonce      []byte `json:"nonce,omitempty"`
	Data       []byte `json:"data"`
}
type Run struct {
	Version    int         `json:"version"`
	ID         string      `json:"id"`
	Kind       string      `json:"kind"`
	InputHash  string      `json:"inputHash"`
	AccountRef string      `json:"accountRef"`
	Source     fd.Source   `json:"source,omitempty"`
	Sources    []fd.Source `json:"sources,omitempty"`
	Status     string      `json:"status"`
	Recurring  bool        `json:"recurring,omitempty"`
	Stage      string      `json:"stage,omitempty"`
	RemoteID   string      `json:"remoteId,omitempty"`
	RemoteIDs  []string    `json:"remoteIds,omitempty"`
	// Candidates from malformed responses are evidence, not confirmed acceptance.
	CandidateRemoteIDs []string        `json:"candidateRemoteIds,omitempty"`
	InputFileHash      string          `json:"inputFileHash,omitempty"`
	Input              json.RawMessage `json:"input,omitempty"`
	Usage              json.RawMessage `json:"usage,omitempty"`
	Receipt            json.RawMessage `json:"receipt,omitempty"`
	ReceiptAccept      string          `json:"receiptAccept,omitempty"`
	ResponseAccept     string          `json:"responseAccept,omitempty"`
	RespondWith        string          `json:"respondWith,omitempty"`
	ResultFormat       string          `json:"resultFormat,omitempty"`
	ItemFailures       int             `json:"itemFailures,omitempty"`
	ErrorCode          string          `json:"errorCode,omitempty"`
	CreatedAt          string          `json:"createdAt"`
	UpdatedAt          string          `json:"updatedAt"`
	ResultHash         string          `json:"resultHash,omitempty"`
}

func ValidRunID(id string) bool { return validRun.MatchString(id) }
func Digest(b []byte) string    { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func New(root, profile, mode string, interactive bool) (*Store, error) {
	if !validName.MatchString(profile) {
		return nil, errors.New("invalid_profile")
	}
	if mode != "auto" && mode != "os" && mode != "file" {
		return nil, errors.New("invalid_credential_store")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, ErrUnsafe
	}
	if info, e := os.Lstat(absolute); e == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, ErrUnsafe
	}
	base := absolute
	suffix := []string{}
	for {
		if _, e := os.Lstat(base); e == nil {
			break
		} else if !os.IsNotExist(e) {
			return nil, ErrUnsafe
		}
		suffix = append([]string{filepath.Base(base)}, suffix...)
		parent := filepath.Dir(base)
		if parent == base {
			return nil, ErrUnsafe
		}
		base = parent
	}
	resolved, err := filepath.EvalSymlinks(base)
	if err != nil {
		return nil, ErrUnsafe
	}
	absolute = filepath.Join(append([]string{resolved}, suffix...)...)
	return &Store{Root: absolute, Dir: filepath.Join(absolute, "profiles", profile), Profile: profile, Mode: mode, Interactive: interactive}, nil
}
func DefaultRoot() (string, error) {
	if runtime.GOOS == "windows" {
		base, err := os.UserConfigDir()
		return filepath.Join(base, "FlashData"), err
	}
	if runtime.GOOS == "darwin" {
		base, err := os.UserConfigDir()
		return filepath.Join(base, "FlashData"), err
	}
	if p := os.Getenv("XDG_STATE_HOME"); filepath.IsAbs(p) {
		return filepath.Join(p, "flashdata"), nil
	}
	h, err := os.UserHomeDir()
	return filepath.Join(h, ".local", "state", "flashdata"), err
}
func (s *Store) Ensure() error {
	for _, p := range []string{s.Root, filepath.Join(s.Root, "profiles"), s.Dir, filepath.Join(s.Dir, "runs"), filepath.Join(s.Dir, "results")} {
		if err := privateDirectory(p); err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) Lock(ctx context.Context) (func(), error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	return lock(ctx, filepath.Join(s.Dir, "profile.lock"))
}
func safeFile(path string, max int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > max || privateCheck(path, info) != nil {
		return nil, ErrUnsafe
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, ErrUnsafe
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil || int64(len(b)) > max {
		return nil, ErrUnsafe
	}
	return b, nil
}
func ReadPrivate(path string, max int64) ([]byte, error) { return safeFile(path, max) }
func AtomicWrite(path string, b []byte) error {
	dir := filepath.Dir(path)
	if err := privateDirectory(dir); err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() || privateCheck(path, info) != nil {
			return ErrUnsafe
		}
	} else if !os.IsNotExist(err) {
		return ErrUnsafe
	}
	f, err := os.CreateTemp(dir, ".atomic-*")
	if err != nil {
		return ErrUnsafe
	}
	name := f.Name()
	defer os.Remove(name)
	if err = secureCreated(name); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return errors.New("state_write_failed")
	}
	if err = replace(name, path); err != nil {
		return errors.New("state_commit_failed")
	}
	return syncDirectory(dir)
}
func (s *Store) Exists() bool {
	for _, name := range []string{"profile.json", "legacy-runs.json", "migration.json"} {
		_, err := os.Lstat(filepath.Join(s.Dir, name))
		if !os.IsNotExist(err) {
			return true
		}
	}
	for _, directory := range []string{"runs", "results"} {
		entries, err := os.ReadDir(filepath.Join(s.Dir, directory))
		if err == nil && len(entries) > 0 {
			return true
		}
	}
	return false
}

// StartMigration pins an interrupted import to the same source snapshot and
// prevents it from replacing a nonempty profile whose vault went missing.
func (s *Store) StartMigration(sourceDigest string) error {
	path := filepath.Join(s.Dir, "migration.json")
	b, err := safeFile(path, 4096)
	if err == nil {
		var marker struct {
			Version      int    `json:"version"`
			SourceDigest string `json:"sourceDigest"`
		}
		if json.Unmarshal(b, &marker) != nil || marker.Version != 1 || marker.SourceDigest != sourceDigest {
			return ErrCorrupt
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	if s.Exists() {
		return ErrCorrupt
	}
	b, _ = json.Marshal(map[string]any{"version": 1, "sourceVersion": 2, "sourceDigest": sourceDigest})
	return AtomicWrite(path, b)
}
func (s *Store) LegacyConflict(id string) (bool, error) {
	b, err := safeFile(filepath.Join(s.Dir, "legacy-runs.json"), 8<<20)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var keys map[string]json.RawMessage
	if json.Unmarshal(b, &keys) != nil {
		return false, ErrCorrupt
	}
	_, ok := keys[Digest([]byte(id))]
	return ok, nil
}
func (s *Store) Load(ctx context.Context) (*Vault, error) {
	for _, path := range []string{s.Root, filepath.Join(s.Root, "profiles"), s.Dir} {
		if info, e := os.Lstat(path); e == nil {
			if !info.IsDir() || privateCheck(path, info) != nil {
				return nil, ErrUnsafe
			}
		} else if !os.IsNotExist(e) {
			return nil, ErrUnsafe
		}
	}

	if info, err := os.Lstat(s.Dir); err == nil {
		if !info.IsDir() || privateCheck(s.Dir, info) != nil {
			return nil, ErrUnsafe
		}
	} else if !os.IsNotExist(err) {
		return nil, ErrUnsafe
	}
	b, err := safeFile(filepath.Join(s.Dir, "vault.json"), 8<<20)
	if os.IsNotExist(err) {
		if s.Exists() {
			return nil, ErrMissing
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var e envelope
	if json.Unmarshal(b, &e) != nil || e.Version != 1 {
		return nil, ErrCorrupt
	}
	plain, err := s.unprotect(ctx, e)
	if err != nil {
		return nil, err
	}
	var v Vault
	if json.Unmarshal(plain, &v) != nil || v.Version != 1 {
		return nil, ErrCorrupt
	}
	return &v, nil
}
func (s *Store) Save(ctx context.Context, v *Vault) error {
	if err := s.Ensure(); err != nil {
		return err
	}
	v.Version = 1
	stored := *v
	if stored.Phase == "environment" {
		if v.originalCredentials != nil && v.originalPhase != "" {
			stored.Credentials = *v.originalCredentials
			stored.Phase = v.originalPhase
		} else {
			stored.Credentials = fd.AgentCredentials{}
			stored.Phase = "environment-required"
		}
	}
	b, err := json.Marshal(stored)
	if err != nil || len(b) > 6<<20 {
		return errors.New("credential_state_full")
	}
	var previous envelope
	if existing, err := safeFile(filepath.Join(s.Dir, "vault.json"), 8<<20); err == nil {
		if json.Unmarshal(existing, &previous) != nil || previous.Version != 1 {
			return ErrCorrupt
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	mode := previous.Protection
	if mode == "" {
		mode = chooseProtection(s.Mode)
	} else if (s.Mode == "file" && mode != "file") || (s.Mode == "os" && mode == "file") {
		return errors.New("credential_store_change_requires_migration")
	}
	e, err := s.protect(ctx, b, mode, previous.Protection == "")
	if err != nil {
		return err
	}
	encoded, _ := json.Marshal(e)
	// This tombstone is saved first; a missing/corrupt vault never starts a new trial.
	if err = AtomicWrite(filepath.Join(s.Dir, "profile.json"), []byte(`{"version":1,"initialized":true}`)); err != nil {
		return err
	}
	return AtomicWrite(filepath.Join(s.Dir, "vault.json"), encoded)
}
func (s *Store) Protection() (string, error) {
	b, err := safeFile(filepath.Join(s.Dir, "vault.json"), 8<<20)
	if err != nil {
		return "", err
	}
	var e envelope
	if json.Unmarshal(b, &e) != nil || e.Version != 1 {
		return "", ErrCorrupt
	}
	return e.Protection, nil
}
func (s *Store) keyAccount() string {
	return Digest([]byte(filepath.Clean(s.Root) + "\x00" + s.Profile))
}
func (s *Store) protect(ctx context.Context, b []byte, mode string, createKey bool) (envelope, error) {
	e := envelope{Version: 1, Protection: mode}
	switch mode {
	case "dpapi":
		data, err := protectOS(b, false)
		e.Data = data
		return e, err
	case "file":
		e.Data = b
		return e, nil
	case "keyring":
		key, err := keyringKey(ctx, s.keyAccount(), createKey, s.Interactive)
		if err != nil {
			return e, ErrVault
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return e, ErrVault
		}
		gcm, _ := cipher.NewGCM(block)
		e.Nonce = make([]byte, gcm.NonceSize())
		if _, err = rand.Read(e.Nonce); err != nil {
			return e, err
		}
		e.Data = gcm.Seal(nil, e.Nonce, b, []byte(s.keyAccount()))
		return e, nil
	default:
		return e, ErrVault
	}
}
func (s *Store) unprotect(ctx context.Context, e envelope) ([]byte, error) {
	switch e.Protection {
	case "dpapi":
		return protectOS(e.Data, true)
	case "file":
		if s.Mode == "os" {
			return nil, ErrVault
		}
		return e.Data, nil
	case "keyring":
		key, err := keyringKey(ctx, s.keyAccount(), false, s.Interactive)
		if err != nil {
			return nil, ErrVault
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, ErrVault
		}
		gcm, _ := cipher.NewGCM(block)
		if len(e.Nonce) != gcm.NonceSize() {
			return nil, ErrCorrupt
		}
		b, err := gcm.Open(nil, e.Nonce, e.Data, []byte(s.keyAccount()))
		if err != nil {
			return nil, ErrCorrupt
		}
		return b, nil
	default:
		return nil, ErrCorrupt
	}
}
func (s *Store) RunPath(id string) string {
	return filepath.Join(s.Dir, "runs", Digest([]byte(id))+".json")
}
func (s *Store) ResultPath(id string) string {
	return filepath.Join(s.Dir, "results", Digest([]byte(id))+".json")
}

func (s *Store) resultVersionDir(id string) string {
	return filepath.Join(s.Dir, "results", Digest([]byte(id)))
}

// The committed hash selects an immutable response version. Older journals and
// bulk streams use ResultPath; a staged response must never replace that file.
func (s *Store) savedResultPath(r *Run) (string, error) {
	if !validResultHash.MatchString(r.ResultHash) {
		return "", ErrCorrupt
	}
	dir := s.resultVersionDir(r.ID)
	info, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return s.ResultPath(r.ID), nil
	}
	if err != nil || !info.IsDir() || privateDirectory(dir) != nil {
		return "", ErrUnsafe
	}
	path := filepath.Join(dir, r.ResultHash+".json")
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return s.ResultPath(r.ID), nil
	} else if err != nil {
		return "", ErrUnsafe
	}
	return path, nil
}

const maxRunBytes = 64 << 20

func (s *Store) ReadRun(id string) (*Run, error) {
	if !ValidRunID(id) {
		return nil, errors.New("invalid_run_id")
	}
	b, err := safeFile(s.RunPath(id), maxRunBytes)
	if os.IsNotExist(err) {
		if _, e := os.Lstat(s.ResultPath(id)); !os.IsNotExist(e) {
			return nil, ErrCorrupt
		}
		if _, e := os.Lstat(s.resultVersionDir(id)); !os.IsNotExist(e) {
			return nil, ErrCorrupt
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var r Run
	if json.Unmarshal(b, &r) != nil || r.Version != 1 || r.ID != id {
		return nil, ErrCorrupt
	}
	return &r, nil
}
func (s *Store) SaveRun(r *Run) error {
	if !ValidRunID(r.ID) {
		return errors.New("invalid_run_id")
	}
	r.Version = 1
	r.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if r.CreatedAt == "" {
		r.CreatedAt = r.UpdatedAt
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if len(b) > maxRunBytes {
		return errors.New("run_record_too_large; no submission may be made")
	}
	return AtomicWrite(s.RunPath(r.ID), b)
}
func (s *Store) SaveResult(r *Run, b []byte) error {
	if !ValidRunID(r.ID) {
		return errors.New("invalid_run_id")
	}
	next := *r
	next.ResultHash = Digest(b)
	next.Status = "succeeded"
	next.ErrorCode = ""
	path := filepath.Join(s.resultVersionDir(r.ID), next.ResultHash+".json")
	if err := AtomicWrite(path, b); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(s.resultVersionDir(r.ID))); err != nil {
		return err
	}
	// Keep staged and previous versions if the journal commit is interrupted.
	// Even an error after replacing the journal can mean it references next.
	if err := s.SaveRun(&next); err != nil {
		return err
	}
	*r = next
	return nil
}
func (s *Store) ReadResult(r *Run) (json.RawMessage, error) {
	path, err := s.savedResultPath(r)
	if err != nil {
		return nil, err
	}
	b, err := safeFile(path, 256<<20)
	if err != nil {
		return nil, errors.New("saved_result_unavailable; preserve the run record")
	}
	if r.ResultHash == "" || Digest(b) != r.ResultHash {
		return nil, ErrCorrupt
	}
	return b, nil
}

// SaveStream retains even a partial export before returning its production error.
func (s *Store) SaveStream(r *Run, produce func(io.Writer) error) error {
	dir := filepath.Join(s.Dir, "results")
	if err := privateDirectory(dir); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".stream-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = secureCreated(name); err != nil {
		f.Close()
		return err
	}
	h := sha256.New()
	productionErr := produce(io.MultiWriter(f, h))
	syncErr := f.Sync()
	closeErr := f.Close()
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err = replace(name, s.ResultPath(r.ID)); err != nil {
		return err
	}
	if err = syncDirectory(dir); err != nil {
		return err
	}
	r.ResultHash = hex.EncodeToString(h.Sum(nil))
	r.ResultFormat = "ndjson"
	r.Status = "succeeded"
	r.ErrorCode = ""
	if productionErr != nil {
		r.Status = "partial"
		r.ErrorCode = "partial_result"
	}
	if err = s.SaveRun(r); err != nil {
		return err
	}
	return productionErr
}
func (s *Store) OpenResult(r *Run) (*os.File, error) {
	path, err := s.savedResultPath(r)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || privateCheck(path, info) != nil {
		return nil, ErrUnsafe
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		f.Close()
		return nil, err
	}
	if r.ResultHash == "" || hex.EncodeToString(h.Sum(nil)) != r.ResultHash {
		f.Close()
		return nil, ErrCorrupt
	}
	if _, err = f.Seek(0, 0); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
func (s *Store) Runs() ([]Run, error) {
	entries, err := os.ReadDir(filepath.Join(s.Dir, "runs"))
	if os.IsNotExist(err) {
		return []Run{}, nil
	}
	if err != nil {
		return nil, ErrUnsafe
	}
	runs := []Run{}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		b, err := safeFile(filepath.Join(s.Dir, "runs", entry.Name()), maxRunBytes)
		if err != nil {
			return nil, err
		}
		var r Run
		if json.Unmarshal(b, &r) != nil || r.Version != 1 || !ValidRunID(r.ID) || entry.Name() != filepath.Base(s.RunPath(r.ID)) {
			return nil, ErrCorrupt
		}
		r.Input = nil
		r.Receipt = nil
		runs = append(runs, r)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].CreatedAt > runs[j].CreatedAt })
	return runs, nil
}
func DecodeLegacy(ctx context.Context, path string) (map[string]json.RawMessage, error) {
	v, _, err := ReadLegacySnapshot(ctx, path)
	return v, err
}

func ReadLegacySnapshot(ctx context.Context, path string) (map[string]json.RawMessage, string, error) {
	b, err := safeFile(path, 4<<20)
	if err != nil {
		return nil, "", err
	}
	v, err := decodeLegacy(b)
	return v, Digest(b), err
}

func decodeLegacy(b []byte) (map[string]json.RawMessage, error) {
	var err error
	var e envelope
	if json.Unmarshal(b, &e) != nil || e.Version != 1 {
		return nil, ErrCorrupt
	}
	var raw []byte
	if e.Protection == "dpapi" {
		raw, err = protectOS(e.Data, true)
	} else if e.Protection == "posix-private-file" && runtime.GOOS != "windows" {
		raw = e.Data
	} else {
		return nil, ErrCorrupt
	}
	if err != nil {
		return nil, err
	}
	var v map[string]json.RawMessage
	if json.Unmarshal(raw, &v) != nil || string(v["version"]) != "2" {
		return nil, ErrCorrupt
	}
	return v, nil
}
func RandomKey() ([]byte, error) { b := make([]byte, 32); _, err := rand.Read(b); return b, err }
func EncodeKey(b []byte) string  { return base64.RawURLEncoding.EncodeToString(b) }
func DecodeKey(s string) ([]byte, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil || len(b) != 32 {
		return nil, ErrVault
	}
	return b, nil
}
