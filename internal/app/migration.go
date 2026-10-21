package app

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
)

func (a *App) addMigration(auth *cobra.Command) {
	var from string
	migrate := &cobra.Command{Use: "migrate", Short: "Explicitly migrate the version-2 script state without resetting history", Args: cobra.NoArgs}
	migrate.Flags().StringVar(&from, "from", "", "legacy private connection.json file")
	migrate.RunE = func(cmd *cobra.Command, args []string) error {
		if from == "" {
			return fail("migration_source_required", "Pass the legacy connection.json path using --from.", "")
		}
		ctx, cancel := a.context(cmd)
		defer cancel()
		legacy, sourceDigest, err := state.ReadLegacySnapshot(ctx, from)
		if err != nil {
			return err
		}
		var origin struct {
			API          string                  `json:"apiOrigin"`
			Data         string                  `json:"dataOrigin"`
			Web          string                  `json:"webOrigin"`
			Phase        string                  `json:"phase"`
			Proof        string                  `json:"registrationProof"`
			Input        *fd.RegistrationRequest `json:"input"`
			Credentials  fd.AgentCredentials     `json:"credentials"`
			ConnectionID string                  `json:"connectionId"`
			Claim        *fd.AgentClaim          `json:"claim"`
		}
		raw, _ := json.Marshal(legacy)
		if json.Unmarshal(raw, &origin) != nil {
			return state.ErrCorrupt
		}
		v := defaultVault()
		if origin.API != "" {
			v.ManagementOrigin, err = fd.NormalizeOrigin(origin.API)
			if err != nil {
				return err
			}
		}
		if origin.Data != "" {
			v.DataOrigin, err = fd.NormalizeOrigin(origin.Data)
			if err != nil {
				return err
			}
		}
		if origin.Web != "" {
			v.WebOrigin, err = fd.NormalizeOrigin(origin.Web)
			if err != nil {
				return err
			}
		}
		v.Phase = origin.Phase
		if v.Phase == "" && origin.Proof != "" {
			v.Phase = "registering"
		}
		v.RegistrationProof = origin.Proof
		v.Registration = origin.Input
		v.Credentials = origin.Credentials
		if v.Credentials.ConnectionID == "" {
			v.Credentials.ConnectionID = origin.ConnectionID
		}
		v.Claim = origin.Claim
		v.MigratedFrom = sourceDigest
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		old, err := s.Load(ctx)
		if err != nil && !errors.Is(err, state.ErrMissing) {
			return err
		}
		if old != nil && old.MigratedFrom != v.MigratedFrom {
			return fail("migration_target_not_empty", "Choose a fresh profile; existing credentials/history were not overwritten.", "")
		}
		if old != nil {
			return a.success(map[string]any{"migrated": true, "alreadyMigrated": true, "sourceVersion": 2, "targetProfileVersion": 1, "sourceFileUnchanged": true, "currentCredentialsRetained": true, "nextCommand": "flashdata auth status"}, "", false, nil)
		}
		if err = s.StartMigration(v.MigratedFrom); err != nil {
			return fail("migration_target_not_empty", "The target has unrelated history or another migration snapshot; it was not overwritten.", "Use a fresh profile or resume with the same original source file.")
		}
		var operations map[string]json.RawMessage
		if len(legacy["operations"]) > 0 && json.Unmarshal(legacy["operations"], &operations) != nil {
			return state.ErrCorrupt
		}
		if operations == nil {
			operations = map[string]json.RawMessage{}
		}
		tombstones := map[string]any{}
		for key := range operations {
			if len(key) != 64 || strings.Trim(key, "0123456789abcdef") != "" {
				return state.ErrCorrupt
			}
			tombstones[key] = map[string]any{"sourceVersion": 2, "record": "legacy-" + key}
		}
		barrier, _ := json.Marshal(tombstones)
		if err = state.AtomicWrite(filepath.Join(s.Dir, "legacy-runs.json"), barrier); err != nil {
			return err
		}
		var accessRef string
		if v.Credentials.APIKey != "" {
			if access, e := a.authorize(ctx, v, "", false); e == nil {
				accessRef = access.AccountRef + "@" + state.Digest([]byte(authOrigin(access)))
			}
		}
		for key, value := range operations {
			id := "legacy-" + key
			if existing, e := s.ReadRun(id); e != nil {
				return e
			} else if existing != nil {
				continue
			}
			var op struct {
				Hash   string          `json:"hash"`
				State  string          `json:"state"`
				Result json.RawMessage `json:"result"`
				Error  string          `json:"error"`
			}
			if json.Unmarshal(value, &op) != nil {
				return state.ErrCorrupt
			}
			r := &state.Run{ID: id, Kind: "legacy:query", InputHash: op.Hash, AccountRef: accessRef, Status: "uncertain", ErrorCode: "legacy_history_requires_inspection", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
			if len(op.Result) > 0 {
				var result struct {
					Source fd.Source       `json:"source"`
					Usage  json.RawMessage `json:"usage"`
				}
				_ = json.Unmarshal(op.Result, &result)
				r.Source = result.Source
				r.Sources = []fd.Source{result.Source}
				r.Usage = result.Usage
			}
			if err = s.SaveRun(r); err != nil {
				return err
			}
			if op.State == "completed" && len(op.Result) > 0 {
				b, _ := json.Marshal(sanitize(op.Result))
				if err = s.SaveResult(r, b); err != nil {
					return err
				}
			}
		}
		if err = s.Save(ctx, v); err != nil {
			return err
		}
		return a.success(map[string]any{"migrated": true, "sourceVersion": 2, "targetProfileVersion": 1, "operationRecords": len(operations), "sourceFileUnchanged": true, "automaticTrialCreated": false, "nextCommand": "flashdata auth status", "note": "Legacy original IDs were stored only as hashes. Their tombstones block automatic replay; inspect legacy-* records explicitly."}, "", false, nil)
	}
	auth.AddCommand(migrate)
}
