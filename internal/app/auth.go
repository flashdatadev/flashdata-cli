package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func (a *App) addAuth() {
	var name, audience, client, scopes string
	init := &cobra.Command{Use: "init", Short: "Reuse saved access or durably register one temporary connection", Args: cobra.NoArgs}
	init.Flags().StringVar(&name, "name", "", "connection name; required to start a new trial")
	init.Flags().StringVar(&audience, "audience", "coding", "coding or personal")
	init.Flags().StringVar(&client, "client", "flashdata-cli", "self-reported client label")
	init.Flags().StringVar(&scopes, "scopes", "google_search,youtube_metadata", "comma-separated trial source names")
	init.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		v, err := s.Load(ctx)
		if err != nil {
			return err
		}
		if os.Getenv("FLASHDATA_API_KEY") != "" || (v != nil && v.Credentials.APIKey != "") {
			v, err = a.load(ctx, s)
			if err != nil {
				return err
			}
			auth, err := a.authorize(ctx, v, "", false)
			if err != nil {
				return err
			}
			return a.success(map[string]any{"status": "reused", "identity": publicIdentity(auth), "nextCommand": "flashdata sources list"}, "", false, nil)
		}
		if v != nil && (v.Phase != "registering" || v.RegistrationProof == "" || v.Registration == nil) {
			return fail("existing_profile_requires_recovery", "This profile already has history or unusable credentials; no replacement trial was created.", "Run auth status, auth migrate, or auth login.")
		}
		if v == nil {
			if _, e := os.Lstat(filepath.Join(s.Root, "connection.json")); e == nil {
				return fail("migration_required", "Existing client state was found.", "Use auth migrate --from PATH; automatic replacement is disabled.")
			}
			if name == "" {
				return fail("name_required", "Starting a trial requires an explicit --name.", "Use init --name NAME --scopes SOURCE,...")
			}
			v, err = a.initialVault()
			if err != nil {
				return err
			}
			c, err := a.client(v, "")
			if err != nil {
				return err
			}
			caps, err := c.Onboarding.Capabilities(ctx)
			if err != nil {
				return err
			}
			if !caps.Available || caps.ProtocolVersion != 2 {
				return fail("onboarding_unavailable", "Protocol 2 registration is not available.", "Import an existing key with auth login.")
			}
			proof, err := fd.NewRegistrationProof()
			if err != nil {
				return err
			}
			requested := []fd.Source{}
			for _, s := range strings.Split(scopes, ",") {
				requested = append(requested, fd.Source(strings.TrimSpace(s)))
			}
			v.RegistrationProof = proof
			v.Registration = &fd.RegistrationRequest{Name: name, Audience: audience, Client: client, Transport: "cli", Scopes: requested}
			v.Phase = "registering"
			if err = fd.ValidateRegistration(*v.Registration); err != nil {
				return err
			}
			if err = s.Save(ctx, v); err != nil {
				return err
			}
		} else if (cmd.Flags().Changed("name") && name != v.Registration.Name) || (cmd.Flags().Changed("scopes") && normalizeScopesFlag(scopes) != joinSources(v.Registration.Scopes)) || (cmd.Flags().Changed("audience") && audience != v.Registration.Audience) || (cmd.Flags().Changed("client") && client != v.Registration.Client) {
			return fail("registration_conflict", "Recover using the exact saved registration input.", "Run init without changing registration flags.")
		}
		c, err := a.client(v, "")
		if err != nil {
			return err
		}
		registration, err := c.Onboarding.Register(ctx, v.RegistrationProof, *v.Registration)
		if err != nil {
			return &Error{Code: "registration_failed", Message: "Registration did not complete; its proof is retained.", Next: "Run init again to recover with the same proof and request.", Cause: err}
		}
		origin, err := fd.NormalizeOrigin(registration.DataOrigin)
		if err != nil || origin != v.DataOrigin {
			return fail("data_origin_mismatch", "The registration returned an unexpected data origin; its proof is retained.", "Inspect server configuration; do not create another trial.")
		}
		v.Credentials = registration.Credentials
		v.Phase = "trial"
		if err = s.Save(ctx, v); err != nil {
			return err
		}
		protection, _ := s.Protection()
		balanceClient, err := a.client(v, v.Credentials.APIKey)
		if err != nil {
			return err
		}
		balance, balanceErr := balanceClient.Account.Balance(ctx)
		return a.success(map[string]any{"status": "trial", "connection": registration.Connection, "startingAllowanceCredits": registration.Connection.TrialCredits, "startingAllowanceRequestEquivalents": registration.Connection.TrialRequestCounts, "sharedAllowanceNote": "Each source count is an alternative use of the same starting grant, not an independent quota or current balance.", "currentBalance": balance, "currentBalanceAvailable": balanceErr == nil, "deniedScopes": registration.DeniedScopes, "credentialProtection": protection, "nextCommand": "flashdata auth claim"}, "", false, nil)
	}
	a.Root.AddCommand(init)
	auth := &cobra.Command{Use: "auth", Short: "Manage API-key profiles and browser claiming"}
	a.Root.AddCommand(auth)
	var stdin, replace bool
	login := &cobra.Command{Use: "login", Short: "Import and verify an API key using hidden input or stdin", Args: cobra.NoArgs}
	login.Flags().BoolVar(&stdin, "api-key-stdin", false, "read exactly one API key from standard input")
	login.Flags().BoolVar(&replace, "replace", false, "explicitly replace saved onboarding credentials, retaining history")
	login.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		var b []byte
		var err error
		if stdin {
			b, err = a.readInput(cmd, "-", 512)
			if err != nil {
				return err
			}
			if len(b) > 512 {
				return fail("invalid_api_key", "Provide one API key on stdin.", "")
			}
		} else {
			if !a.prompting() {
				return fail("stdin_required", "Noninteractive login requires --api-key-stdin.", "Pipe the key from a secret manager; never put it in arguments.")
			}
			f, ok := a.IO.In.(*os.File)
			if !ok {
				return fail("terminal_required", "Hidden input requires a terminal.", "")
			}
			_, _ = fmt.Fprint(a.IO.Err, "API key: ")
			b, err = term.ReadPassword(int(f.Fd()))
			_, _ = fmt.Fprintln(a.IO.Err)
			if err != nil {
				return fail("key_input_failed", "Could not read hidden input.", "")
			}
		}
		key := strings.TrimSpace(string(b))
		for i := range b {
			b[i] = 0
		}
		if !strings.HasPrefix(key, "fd_") || strings.ContainsAny(key, " \r\n\t") {
			return fail("invalid_api_key", "Provide an ordinary FlashData API key; MCP OAuth tokens cannot be used here.", "")
		}
		a.knownSecrets = append(a.knownSecrets, key)
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		old, err := s.Load(ctx)
		if err != nil && !errors.Is(err, state.ErrMissing) {
			return err
		}
		if old != nil && old.Phase != "logged_out" && old.Phase != "environment-required" && old.Credentials.APIKey != key && !replace {
			return fail("replacement_requires_flag", "This profile already has saved access.", "Use a new --profile or auth login --replace --api-key-stdin.")
		}
		if old != nil && old.Credentials.APIKey == key && !replace {
			verified, err := a.authorize(ctx, old, "", false)
			if err != nil {
				return err
			}
			protection, _ := s.Protection()
			return a.success(map[string]any{"status": "reused", "identity": publicIdentity(verified), "credentialProtection": protection, "recoveryStateRetained": true}, "", false, nil)
		}
		v := old
		if v == nil {
			v, err = a.initialVault()
			if err != nil {
				return err
			}
		}
		candidate := *v
		candidate.Credentials = fd.AgentCredentials{APIKey: key}
		candidate.Phase = "imported"
		candidate.Registration = nil
		candidate.RegistrationProof = ""
		candidate.Claim = nil
		if replace || old == nil || old.Phase != "environment-required" {
			candidate.Uploads = nil
			candidate.BulkRuns = nil
		}
		verified, err := a.authorize(ctx, &candidate, "", false)
		if err != nil {
			return err
		}
		if err = s.Save(ctx, &candidate); err != nil {
			return err
		}
		protection, _ := s.Protection()
		return a.success(map[string]any{"status": "imported", "identity": publicIdentity(verified), "credentialProtection": protection}, "", false, nil)
	}
	auth.AddCommand(login)
	status := &cobra.Command{Use: "status", Short: "Verify access, scopes, expiry and claim status", Args: cobra.NoArgs}
	status.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, err := a.store()
		if err != nil {
			return err
		}
		v, err := a.readProfile(ctx, s)
		if err != nil {
			return err
		}
		if v.Credentials.APIKey == "" {
			if v.Phase == "" {
				return state.ErrMissing
			}
			next := "flashdata auth login --replace"
			if v.Phase == "registering" {
				next = "flashdata init"
			}
			protection, _ := s.Protection()
			return a.success(map[string]any{"profile": s.Profile, "phase": v.Phase, "accessVerified": false, "credentialProtection": protection, "nextCommand": next}, "", false, nil)
		}
		access, verifyErr := a.authorize(ctx, v, "", false)
		if verifyErr != nil && v.Credentials.HandoffProof == "" {
			return verifyErr
		}
		var claim any
		if v.Credentials.HandoffProof != "" {
			c, err := a.client(v, "")
			if err != nil {
				return err
			}
			claim, err = c.Onboarding.Status(ctx, v.Credentials.ConnectionID, v.Credentials.HandoffProof)
			if err != nil && verifyErr != nil {
				return verifyErr
			}
		}
		protection, _ := s.Protection()
		return a.success(map[string]any{"profile": s.Profile, "phase": v.Phase, "identity": publicIdentity(access), "accessVerified": verifyErr == nil, "claim": claim, "credentialProtection": protection}, "", false, nil)
	}
	auth.AddCommand(status)
	claim := &cobra.Command{Use: "claim", Short: "Create the browser approval URL and separate confirmation code", Args: cobra.NoArgs}
	claim.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		v, err := s.Load(ctx)
		if err != nil {
			return err
		}
		if v == nil || v.Credentials.SetupCredential == "" {
			return fail("claim_unavailable", "This profile has no temporary setup credential.", "Use init for a new explicit connection, or auth finish for a pending claim.")
		}
		c, err := a.client(v, "")
		if err != nil {
			return err
		}
		result, err := c.Onboarding.CreateClaim(ctx, v.Credentials.ConnectionID, v.Credentials.SetupCredential)
		if err != nil {
			return err
		}
		v.Claim = result
		if err = s.Save(ctx, v); err != nil {
			return err
		}
		url, err := fd.ClaimURL(v.WebOrigin, result)
		if err != nil {
			return err
		}
		return a.success(map[string]any{"claimUrl": url, "confirmationCode": result.ConfirmationCode, "expiresAt": result.ExpiresAt, "nextCommand": "flashdata auth finish --wait"}, "", false, nil)
	}
	auth.AddCommand(claim)
	var wait bool
	finish := &cobra.Command{Use: "finish", Short: "Save, independently verify and acknowledge the approved owner key", Args: cobra.NoArgs}
	finish.Flags().BoolVar(&wait, "wait", false, "wait for browser approval within --timeout")
	finish.RunE = func(cmd *cobra.Command, args []string) error { return a.finish(cmd, wait) }
	auth.AddCommand(finish)
	logout := &cobra.Command{Use: "logout", Short: "Remove local credentials and retain run history; does not revoke remote access", Args: cobra.NoArgs}
	logout.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		v, err := s.Load(ctx)
		if err != nil {
			return err
		}
		if v == nil {
			v, err = a.initialVault()
			if err != nil {
				return err
			}
		}
		v.Credentials = fd.AgentCredentials{}
		v.RegistrationProof = ""
		v.Registration = nil
		v.Claim = nil
		v.Uploads = nil
		v.Phase = "logged_out"
		if err = s.Save(ctx, v); err != nil {
			return err
		}
		return a.success(map[string]any{"status": "logged_out", "historyRetained": true, "remoteAccessRevoked": false, "environmentKeyStillSet": os.Getenv("FLASHDATA_API_KEY") != "", "nextAction": "Unset FLASHDATA_API_KEY separately if configured; revoke remote keys in your FlashData account when needed."}, "", false, nil)
	}
	auth.AddCommand(logout)
	verify := &cobra.Command{Use: "__verify-install", Hidden: true, Args: cobra.NoArgs}
	verify.RunE = func(cmd *cobra.Command, args []string) error { return a.verifyInstall(cmd) }
	a.Root.AddCommand(verify)
	a.addMigration(auth)
}
func joinSources(s []fd.Source) string {
	v := []string{}
	for _, x := range s {
		v = append(v, string(x))
	}
	return strings.Join(v, ",")
}

func normalizeScopesFlag(value string) string {
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, ",")
}
func publicIdentity(access *access) any {
	if access == nil {
		return nil
	}
	id := *access.Identity
	if access.Method == "key_access_only" {
		id.AccountID = ""
	}
	return map[string]any{"method": access.Method, "identity": id, "connection": access.Connection, "permittedSources": access.Sources}
}
func (a *App) finish(cmd *cobra.Command, wait bool) error {
	ctx, cancel := a.context(cmd)
	defer cancel()
	s, unlock, err := a.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	v, err := s.Load(ctx)
	if err != nil {
		return err
	}
	if v == nil {
		return state.ErrMissing
	}
	if v.Phase == "installed" {
		auth, err := a.authorize(ctx, v, "", false)
		if err != nil {
			return err
		}
		return a.success(map[string]any{"status": "installed", "identity": publicIdentity(auth)}, "", false, nil)
	}
	if v.Credentials.HandoffProof == "" {
		return fail("handoff_proof_missing", "No recoverable handoff proof exists.", "Use auth status; do not create replacement trials.")
	}
	c, err := a.client(v, "")
	if err != nil {
		return err
	}
	for v.Phase != "installing" {
		status, err := c.Onboarding.Status(ctx, v.Credentials.ConnectionID, v.Credentials.HandoffProof)
		if err != nil {
			return err
		}
		if status.Status == "handoff_ready" {
			handoff, err := c.Onboarding.Handoff(ctx, v.Credentials.ConnectionID, v.Credentials.HandoffProof)
			if err != nil {
				return err
			}
			if handoff.Status != "delivered" || handoff.Credentials == nil {
				return fail("reconnect_required", "The handoff has already been confirmed without a local candidate.", "Recover your installed key or reconnect through your FlashData account.")
			}
			proof := v.Credentials.HandoffProof
			v.Credentials = *handoff.Credentials
			v.Credentials.HandoffProof = proof
			v.Phase = "installing"
			if err = s.Save(ctx, v); err != nil {
				return err
			}
			break
		}
		if status.Status != "awaiting_claim" && status.Status != "claim_processing" {
			return fail("onboarding_"+status.Status, "The server requires recovery: "+status.NextAction, "Inspect auth status and the browser account; no new trial was created.")
		}
		if !wait {
			return fail("approval_pending", "Browser approval is still pending.", "Finish browser approval, then run auth finish --wait.")
		}
		seconds := status.PollAfterSeconds
		if seconds < 1 {
			seconds = 5
		}
		if seconds > 60 {
			seconds = 60
		}
		timer := time.NewTimer(time.Duration(seconds) * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	verifier := a.VerifyFresh
	if verifier == nil {
		verifier = a.freshProcess
	}
	if err = verifier(ctx, s, v); err != nil {
		return fail("credential_installation_unverified", "The saved candidate was not independently verified; no ACK was sent.", "Keep the saved key/proof and run auth finish again.")
	}
	if _, err = c.Onboarding.Acknowledge(ctx, v.Credentials.ConnectionID, v.Credentials.HandoffProof, v.Credentials.APIKey); err != nil {
		return err
	}
	loadedClient, err := a.client(v, v.Credentials.APIKey)
	if err != nil {
		return err
	}
	identity, err := loadedClient.Onboarding.Identity(ctx, v.Credentials.ConnectionID, v.Credentials.APIKey)
	if err != nil {
		return err
	}
	if !identity.Installed || identity.Ownership != "owned" || identity.Generation != v.Credentials.Generation {
		return fail("installation_confirmation_missing", "ACK completed but installed identity is not confirmed.", "Retain state and run auth finish again.")
	}
	v.Phase = "installed"
	v.Credentials.HandoffProof = ""
	v.Credentials.SetupCredential = ""
	v.RegistrationProof = ""
	v.Registration = nil
	v.Claim = nil
	if err = s.Save(ctx, v); err != nil {
		return err
	}
	return a.success(map[string]any{"status": "installed", "connection": identity, "nextCommand": "flashdata auth status"}, "", false, nil)
}

type installProof struct {
	ConnectionID string `json:"connectionId"`
	Generation   int    `json:"generation"`
	Nonce        string `json:"nonce"`
}

func (a *App) verifyInstall(cmd *cobra.Command) error {
	ctx, cancel := a.context(cmd)
	defer cancel()
	s, err := a.store()
	if err != nil {
		return err
	}
	v, err := s.Load(ctx)
	if err != nil {
		return err
	}
	if v == nil || (v.Phase != "installing" && v.Phase != "installed") {
		return state.ErrMissing
	}
	b, err := io.ReadAll(io.LimitReader(a.IO.In, 4097))
	if err != nil || len(b) > 4096 {
		return fail("invalid_verification_challenge", "Invalid fresh-process challenge.", "")
	}
	var challenge installProof
	if decodeStrict(b, &challenge) != nil || challenge.Nonce == "" || challenge.ConnectionID != v.Credentials.ConnectionID || challenge.Generation != v.Credentials.Generation {
		return fail("invalid_verification_challenge", "The loaded candidate does not match the challenge.", "")
	}
	c, err := a.client(v, v.Credentials.APIKey)
	if err != nil {
		return err
	}
	id, err := c.Onboarding.Identity(ctx, v.Credentials.ConnectionID, v.Credentials.APIKey)
	if err != nil {
		return err
	}
	if id.ID != challenge.ConnectionID || id.Generation != challenge.Generation || id.Ownership != "owned" || id.Lifecycle != "valid" || id.Access != "ready" {
		return fail("credential_installation_unverified", "The saved key is not the approved owner candidate.", "")
	}
	return a.success(challenge, "", false, nil)
}
func (a *App) freshProcess(ctx context.Context, s *state.Store, v *state.Vault) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	nonce, err := fd.NewRegistrationProof()
	if err != nil {
		return err
	}
	challenge := installProof{v.Credentials.ConnectionID, v.Credentials.Generation, nonce}
	input, _ := json.Marshal(challenge)
	child := exec.CommandContext(ctx, executable, "__verify-install", "--state-dir", s.Root, "--profile", s.Profile, "--credential-store", s.Mode, "--timeout", a.Timeout.String(), "--json", "--non-interactive")
	child.Stdin = bytes.NewReader(input)
	var output bytes.Buffer
	child.Stdout = &output
	child.Stderr = io.Discard
	for _, e := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(e), "FLASHDATA_API_KEY=") {
			child.Env = append(child.Env, e)
		}
	}
	configureChild(child)
	if child.Run() != nil || output.Len() > 8192 {
		return errors.New("fresh_process_failed")
	}
	var result struct {
		OK   bool         `json:"ok"`
		Data installProof `json:"data"`
	}
	if json.Unmarshal(output.Bytes(), &result) != nil || !result.OK || result.Data != challenge {
		return errors.New("fresh_process_failed")
	}
	return nil
}
