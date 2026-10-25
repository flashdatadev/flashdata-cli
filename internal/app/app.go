package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type IO struct {
	In       io.Reader
	Out, Err io.Writer
}
type App struct {
	cancelCommand context.CancelFunc

	IO                                      IO
	Version, Commit                         string
	Root                                    *cobra.Command
	Profile, StateDir, StoreMode            string
	DataOrigin, ManagementOrigin, WebOrigin string
	JSON, NonInteractive                    bool
	Timeout                                 time.Duration
	HTTPClient                              *http.Client
	VerifyFresh                             func(context.Context, *state.Store, *state.Vault) error
	interactive                             bool
	knownSecrets                            []string
}
type failure struct {
	Code       string `json:"code"`
	Message    string `json:"message,omitempty"`
	RunID      string `json:"runId,omitempty"`
	RemoteID   string `json:"remoteId,omitempty"`
	NextAction string `json:"nextAction,omitempty"`
	HTTPStatus int    `json:"httpStatus,omitempty"`
	RetryAfter string `json:"retryAfter,omitempty"`
	Details    any    `json:"details,omitempty"`
}
type Envelope struct {
	SchemaVersion int             `json:"schemaVersion"`
	OK            bool            `json:"ok"`
	RunID         string          `json:"runId,omitempty"`
	Cached        bool            `json:"cached"`
	Data          any             `json:"data,omitempty"`
	Usage         json.RawMessage `json:"usage,omitempty"`
	Error         *failure        `json:"error,omitempty"`
}
type Error struct {
	Code, Message, Next, RunID, RemoteID string
	Cause                                error
	Details                              any
}

func (e *Error) Error() string              { return e.Code + ": " + e.Message }
func (e *Error) Unwrap() error              { return e.Cause }
func fail(code, message, next string) error { return &Error{Code: code, Message: message, Next: next} }
func New(streams IO, version, commit string) *App {
	a := &App{IO: streams, Version: version, Commit: commit}
	if f, ok := streams.In.(*os.File); ok {
		a.interactive = term.IsTerminal(int(f.Fd()))
	}
	root := &cobra.Command{Use: "flashdata", Short: "FlashData search, YouTube and bulk data workflows", Long: "FlashData CLI — https://flashdata.dev\nStandalone API-key access with secure profiles and durable local run recovery.", SilenceUsage: true, SilenceErrors: true, Args: cobra.NoArgs, CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true}}
	a.Root = root
	root.SetIn(streams.In)
	root.SetOut(streams.Out)
	root.SetErr(streams.Err)
	f := root.PersistentFlags()
	f.StringVar(&a.Profile, "profile", "default", "isolated saved connection")
	f.StringVar(&a.StateDir, "state-dir", "", "private persistent state directory")
	f.StringVar(&a.StoreMode, "credential-store", "auto", "auto, os, or file (protected persistent file)")
	f.BoolVar(&a.JSON, "json", false, "versioned machine-readable output")
	f.BoolVar(&a.NonInteractive, "non-interactive", false, "disable prompts and browser opening")
	f.DurationVar(&a.Timeout, "timeout", 2*time.Minute, "overall command deadline; accepted server work continues")
	f.StringVar(&a.DataOrigin, "data-origin", "", "pin the data origin for a new profile")
	f.StringVar(&a.ManagementOrigin, "management-origin", "", "pin the management origin for a new profile")
	f.StringVar(&a.WebOrigin, "web-origin", "", "pin the browser origin for a new profile")
	var showVersion bool
	root.Flags().BoolVar(&showVersion, "version", false, "show version offline")
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if showVersion {
			return a.success(map[string]string{"version": a.Version, "commit": a.Commit}, "", false, nil)
		}
		return cmd.Help()
	}
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if a.Timeout <= 0 {
			return fail("invalid_timeout", "timeout must be positive", "")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), a.Timeout)
		a.cancelCommand = cancel
		cmd.SetContext(ctx)
		return nil
	}
	a.addAuth()
	a.addAgentSetup()
	a.addQueries()
	a.addAccount()
	a.addBulk()
	a.addRuns()
	a.addDiscovery()
	return a
}
func (a *App) Execute(ctx context.Context, args []string) int {
	defer func() {
		if a.cancelCommand != nil {
			a.cancelCommand()
		}
	}()
	a.Root.SetArgs(args)
	err := a.Root.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	f := a.classify(err)
	if a.machine() {
		_ = json.NewEncoder(a.IO.Err).Encode(Envelope{SchemaVersion: 1, OK: false, Error: f, RunID: f.RunID})
	} else {
		_, _ = fmt.Fprintln(a.IO.Err, f.Code+": "+f.Message)
		if f.RunID != "" {
			_, _ = fmt.Fprintln(a.IO.Err, "Run:", f.RunID)
		}
		if f.NextAction != "" {
			_, _ = fmt.Fprintln(a.IO.Err, f.NextAction)
		}
	}
	return 1
}
func (a *App) machine() bool   { return a.JSON || !a.interactive || a.NonInteractive }
func (a *App) prompting() bool { return a.interactive && !a.NonInteractive }
func (a *App) success(data any, run string, cached bool, usage json.RawMessage) error {
	return a.writeSuccess(a.redacted(data), run, cached, usage)
}

// writeSuccess accepts already-sanitized output, including embedded SDK schemas
// whose credential property definitions are public metadata, not secret values.
func (a *App) writeSuccess(data any, run string, cached bool, usage json.RawMessage) error {
	e := Envelope{SchemaVersion: 1, OK: true, Data: data, RunID: run, Cached: cached, Usage: usage}
	encoder := json.NewEncoder(a.IO.Out)
	encoder.SetEscapeHTML(false)
	if !a.machine() {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(e)
}

var secretPattern = regexp.MustCompile(`(?i)\b(?:fd_|fda_setup_|fda_handoff_)[A-Za-z0-9_-]+`)
var urlSecret = regexp.MustCompile(`https?://[^\s<>"']+\?[^\s<>"']+`)

func (a *App) clean(s string) string {
	for _, v := range a.knownSecrets {
		if v != "" {
			s = strings.ReplaceAll(s, v, "[REDACTED]")
		}
	}
	s = secretPattern.ReplaceAllString(s, "[REDACTED]")
	s = urlSecret.ReplaceAllString(s, "[REDACTED URL]")
	return s
}
func (a *App) classify(err error) *failure {
	f := &failure{Code: "command_failed", Message: "Command failed. Preserve local state and inspect the error."}
	var e *Error
	if errors.As(err, &e) {
		f.Code = e.Code
		f.Message = a.clean(e.Message)
		f.NextAction = e.Next
		f.RunID = e.RunID
		f.RemoteID = e.RemoteID
		f.Details = e.Details
	}
	var api *fd.APIError
	if errors.As(err, &api) {
		if f.Code == "command_failed" {
			f.Code = api.Code
			f.Message = a.clean(api.Message)
		}
		f.HTTPStatus = api.StatusCode
		f.RetryAfter = api.RetryAfter
		if f.NextAction == "" && (api.StatusCode == 401 || api.StatusCode == 403) {
			f.NextAction = "Run auth status or import the intended key; no replacement trial was created."
		}
	}
	var validation *fd.ValidationError
	if errors.As(err, &validation) && f.Code == "command_failed" {
		f.Code = "invalid_input"
		f.Message = a.clean(validation.Error())
	}
	if fd.IsSubmissionUnknown(err) {
		f.Code = "request_outcome_unknown"
		f.Message = "The server may have accepted the request."
		f.NextAction = "Inspect the existing run before submitting again."
		var unknown *fd.SubmissionUnknownError
		if errors.As(err, &unknown) && len(unknown.JobIDs) > 0 {
			f.Details = map[string]any{"candidateRemoteIds": unknown.JobIDs, "acceptanceConfirmed": false}
			f.NextAction = "Inspect candidateRemoteIds with jobs get or bulk get; reconcile the complete submission before creating new work."
		}
	}
	if f.Code == "command_failed" && errors.Is(err, context.DeadlineExceeded) {
		f.Code = "timeout"
		f.Message = "Local waiting timed out; admitted work may continue."
	}
	if f.Code == "command_failed" && errors.Is(err, context.Canceled) {
		f.Code = "cancelled"
		f.Message = "Local work was cancelled; admitted work may continue."
	}
	if f.Code == "command_failed" {
		switch {
		case errors.Is(err, state.ErrUnsafe):
			f.Code = "unsafe_state_storage"
			f.Message = "State permissions, ownership or file type are unsafe."
		case errors.Is(err, state.ErrCorrupt):
			f.Code = "state_corrupt"
			f.Message = "Preserve existing state; automatic reset is disabled."
		case errors.Is(err, state.ErrMissing):
			f.Code = "credentials_missing"
			f.Message = "Import an existing key with auth login."
		case errors.Is(err, state.ErrVault):
			f.Code = "credential_store_unavailable"
			f.Message = "Unlock the original credential store; do not reset or create replacement access."
		case errors.Is(err, state.ErrBusy):
			f.Code = "profile_busy"
			f.Message = "Another process holds this profile lock."
		default:
			f.Message = "Command failed. Check command usage, storage permissions and connectivity."
			if strings.HasPrefix(err.Error(), "unknown command") || strings.HasPrefix(err.Error(), "unknown flag") || strings.HasPrefix(err.Error(), "invalid argument") {
				f.Code = "invalid_arguments"
				f.NextAction = "Use --help for the accepted commands and options; never pass secrets as arguments."
			}
		}
	}
	if f.Code == "" {
		f.Code = "api_request_failed"
	}
	f.NextAction = a.clean(f.NextAction)
	f.Details = a.redacted(f.Details)
	return f
}
func (a *App) context(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return context.WithTimeout(cmd.Context(), a.Timeout)
}
func (a *App) store() (*state.Store, error) {
	root := a.StateDir
	if root == "" {
		var err error
		root, err = state.DefaultRoot()
		if err != nil {
			return nil, err
		}
	}
	return state.New(root, a.Profile, a.StoreMode, a.prompting())
}
func (a *App) lock(ctx context.Context) (*state.Store, func(), error) {
	s, err := a.store()
	if err != nil {
		return nil, nil, err
	}
	unlock, err := s.Lock(ctx)
	return s, unlock, err
}
func defaultVault() *state.Vault {
	return &state.Vault{Version: 1, DataOrigin: "https://data.flashdata.dev", ManagementOrigin: "https://api.flashdata.dev", WebOrigin: "https://flashdata.dev"}
}
func (a *App) initialVault() (*state.Vault, error) {
	v := defaultVault()
	for _, p := range []struct {
		value string
		dest  *string
	}{{a.DataOrigin, &v.DataOrigin}, {a.ManagementOrigin, &v.ManagementOrigin}, {a.WebOrigin, &v.WebOrigin}} {
		if p.value != "" {
			o, err := fd.NormalizeOrigin(p.value)
			if err != nil {
				return nil, err
			}
			*p.dest = o
		}
	}
	return v, nil
}
func (a *App) client(v *state.Vault, key string) (*fd.Client, error) {
	if v == nil {
		var err error
		v, err = a.initialVault()
		if err != nil {
			return nil, err
		}
	}
	for _, p := range []struct{ provided, saved string }{{a.DataOrigin, v.DataOrigin}, {a.ManagementOrigin, v.ManagementOrigin}, {a.WebOrigin, v.WebOrigin}} {
		if p.provided != "" {
			normalized, err := fd.NormalizeOrigin(p.provided)
			if err != nil {
				return nil, err
			}
			if normalized != p.saved {
				return nil, fail("origin_conflict", "Saved profile origins cannot be overridden.", "Use a separate profile for another service origin.")
			}
		}
	}
	if key != "" && !strings.HasPrefix(key, "fd_") {
		return nil, fail("ordinary_api_key_required", "Direct REST commands require an ordinary FlashData API key, not an MCP OAuth token.", "")
	}
	if key != "" {
		a.knownSecrets = append(a.knownSecrets, key)
	}
	return fd.New(fd.Options{APIKey: key, DataOrigin: v.DataOrigin, ManagementOrigin: v.ManagementOrigin, HTTPClient: a.HTTPClient})
}
func (a *App) load(ctx context.Context, s *state.Store) (*state.Vault, error) {
	v, err := a.readProfile(ctx, s)
	if err != nil {
		return nil, err
	}
	if v.Credentials.APIKey == "" {
		return nil, state.ErrMissing
	}
	return v, nil
}

func (a *App) readProfile(ctx context.Context, s *state.Store) (*state.Vault, error) {
	v, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	if v == nil {
		v, err = a.initialVault()
		if err != nil {
			return nil, err
		}
	}
	if key := os.Getenv("FLASHDATA_API_KEY"); key != "" {
		v.UseEnvironmentKey(key)
	}
	a.knownSecrets = append(a.knownSecrets, v.Credentials.APIKey, v.Credentials.SetupCredential, v.Credentials.HandoffProof, v.RegistrationProof)
	return v, nil
}

func requireSource(auth *access, source fd.Source) error {
	if source == "" {
		return nil
	}
	for _, permitted := range auth.Sources {
		if permitted == source {
			return nil
		}
	}
	return fail("source_not_permitted", "The selected key does not permit this source.", "Inspect sources list or import the intended key.")
}

func (a *App) authorizeSources(ctx context.Context, v *state.Vault, sources []fd.Source, recurring bool) (*access, error) {
	auth, err := a.authorize(ctx, v, "", recurring)
	if err != nil {
		return nil, err
	}
	for _, source := range sources {
		if err := requireSource(auth, source); err != nil {
			return nil, err
		}
	}
	return auth, nil
}

type access struct {
	Connection *fd.AgentConnection
	Client     *fd.Client
	Identity   *fd.APIKeyIdentity
	Sources    []fd.Source
	AccountRef string
	Method     string
	Origin     string
}

func (a *App) authorize(ctx context.Context, v *state.Vault, source fd.Source, recurring bool) (*access, error) {
	c, err := a.client(v, v.Credentials.APIKey)
	if err != nil {
		return nil, err
	}
	metadata, err := c.Account.Access(ctx)
	if err != nil {
		return nil, err
	}
	scopes := []fd.Source{}
	for _, s := range metadata.Sources {
		scopes = append(scopes, fd.Source(s))
	}
	id, err := c.Account.Identity(ctx)
	method := "verified_account"
	var verifiedConnection *fd.AgentConnection
	if err != nil {
		var api *fd.APIError
		if !errors.As(err, &api) || api.StatusCode != 404 {
			return nil, err
		}
		if v.Credentials.ConnectionID != "" {
			connection, e := c.Onboarding.Identity(ctx, v.Credentials.ConnectionID, v.Credentials.APIKey)
			if e != nil {
				return nil, e
			}
			if connection.Lifecycle != "valid" || connection.Access != "ready" {
				return nil, fail("access_blocked", "The connection is no longer ready.", "Run auth status.")
			}
			id = &fd.APIKeyIdentity{AccountID: fmt.Sprintf("connection:%s:%s:%d", connection.ID, connection.Ownership, connection.Generation), Ownership: connection.Ownership, Installed: connection.Installed, Access: connection.Access, Scopes: connection.Scopes, RecurringAllowed: connection.Ownership == "owned" && connection.Installed}
			verifiedConnection = connection
			id.AccountType = "owned"
			if connection.Ownership != "owned" {
				id.AccountType = "temporary"
				if connection.TrialExpiresAt != "" {
					id.ExpiresAt = &connection.TrialExpiresAt
				}
			}
			method = "verified_connection"
		} else {
			if recurring {
				return nil, fail("identity_unavailable", "Recurring use needs authoritative owner metadata.", "Deploy the management identity endpoint, then run auth status.")
			}
			id = &fd.APIKeyIdentity{AccountID: "key:" + state.Digest([]byte(v.Credentials.APIKey)), Ownership: "unknown", Access: "ready", Scopes: scopes}
			method = "key_access_only"
		}
	}
	if id.Access != "ready" {
		return nil, fail("access_blocked", "The account/key is blocked.", "Run auth status.")
	}
	allowed := func(s fd.Source) bool {
		inAccess, inID := false, false
		for _, v := range scopes {
			inAccess = inAccess || v == s
		}
		for _, v := range id.Scopes {
			inID = inID || v == s
		}
		return inAccess && inID
	}
	if source != "" && !allowed(source) {
		return nil, fail("source_not_permitted", "The selected key does not permit this source.", "Inspect sources list or import the intended key.")
	}
	if recurring && !id.RecurringAllowed {
		return nil, fail("owned_access_required", "Recurring work requires verified owner access and confirmed installation.", "Complete auth claim and auth finish --wait, or import a verified owner key.")
	}
	filtered := []fd.Source{}
	for _, s := range scopes {
		if allowed(s) {
			filtered = append(filtered, s)
		}
	}
	return &access{Client: c, Identity: id, Connection: verifiedConnection, Sources: filtered, AccountRef: id.AccountID, Method: method, Origin: v.DataOrigin + "|" + v.ManagementOrigin}, nil
}
func readInput(path string, in io.Reader, max int64) ([]byte, error) {
	var reader io.Reader = in
	var f *os.File
	if path != "-" {
		if path == "" {
			return nil, fail("input_required", "Pass --input FILE or --input -.", "")
		}
		var err error
		f, err = os.Open(path)
		if err != nil {
			return nil, fail("input_unreadable", "Could not open the input file.", "")
		}
		defer f.Close()
		reader = f
	}
	b, err := io.ReadAll(io.LimitReader(reader, max+1))
	if err != nil || int64(len(b)) > max {
		return nil, fail("input_too_large", "Input is unreadable or exceeds its size limit.", "")
	}
	return b, nil
}
func decodeStrict(b []byte, out any) error {
	if uniqueJSON(b) != nil {
		return fail("invalid_json", "Provide one JSON value with unique field names and at most 64 nesting levels.", "")
	}
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.UseNumber()
	d.DisallowUnknownFields()
	if d.Decode(out) != nil {
		return fail("invalid_json", "Expected documented JSON fields and value types.", "")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fail("invalid_json", "Provide exactly one JSON value.", "")
	}
	return nil
}
func ensureOutput(path string) error {
	if path == "" || path == "-" {
		return nil
	}
	parent := filepath.Dir(path)
	if info, err := os.Stat(parent); err != nil || !info.IsDir() {
		return fail("output_unwritable", "Output directory does not exist.", "")
	}
	if _, err := os.Lstat(path); err == nil || !os.IsNotExist(err) {
		return fail("output_exists", "Choose a new output file; overwriting is disabled.", "")
	}
	return nil
}
