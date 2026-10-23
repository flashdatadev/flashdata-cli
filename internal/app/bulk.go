package app

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
)

func (a *App) addBulk() {
	bulk := &cobra.Command{Use: "bulk", Short: "Create, upload, control and recover durable bulk tasks"}
	a.Root.AddCommand(bulk)
	var source, input, optionsPath, storagePath string
	f := &queryFlags{}
	run := &cobra.Command{Use: "run", Short: "Create once, upload IDs, start, wait and save NDJSON receipts", Args: cobra.NoArgs}
	bindRun(run, f, false)
	run.Flags().StringVar(&source, "source", "youtube_metadata", "one of five YouTube bulk sources")
	run.Flags().StringVar(&input, "input", "", "UTF-8 video-ID file")
	run.Flags().StringVar(&optionsPath, "options", "", "private shared source-options JSON file")
	run.Flags().StringVar(&storagePath, "storage-config", "", "private customer storage JSON")
	run.RunE = func(cmd *cobra.Command, args []string) error {
		request, err := bulkRequest(fd.Source(source), optionsPath, storagePath)
		a.rememberSecrets(request)
		if err != nil {
			return err
		}
		file, hash, size, err := validateIDs(cmd.Context(), input)
		if err != nil {
			return err
		}
		defer closeIDs(file)
		if err = ensureOutput(f.Output); err != nil {
			return err
		}
		id, err := a.runID(*f)
		if err != nil {
			return err
		}
		canonical, _ := json.Marshal(map[string]any{"request": request, "inputFileHash": hash})
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		v, err := a.load(ctx, s)
		if err != nil {
			return err
		}
		auth, err := a.authorize(ctx, v, request.Source, f.Recurring)
		if err != nil {
			return err
		}
		if f.DryRun {
			pricing, e := auth.Client.Account.Pricing(ctx, fd.PricingParams{Source: request.Source})
			return a.success(map[string]any{"dryRun": true, "request": sanitize(request), "inputBytes": size, "inputFileHash": hash, "pricing": pricing, "pricingAvailable": e == nil, "note": "No job was created. Download billing depends on delivered bytes; per-request data billing depends on accepted IDs."}, "", false, nil)
		}
		r, exists, err := a.begin(ctx, s, v, auth, id, "bulk:run", canonical, []fd.Source{request.Source}, f.Recurring)
		if err != nil {
			return err
		}
		if exists {
			if r.Status == "succeeded" {
				return a.exportRun(s, r, f.Output, true)
			}
			if r.RemoteID == "" && v.BulkRuns[r.ID] != "" {
				r.RemoteID = v.BulkRuns[r.ID]
				r.Status = "accepted"
			}
			return a.recoverRun(ctx, s, auth, r, f.Output)
		}
		r.InputFileHash = hash
		if err = s.SaveRun(r); err != nil {
			return err
		}
		created, err := auth.Client.Bulk.Create(ctx, request)
		if err != nil {
			return a.failureRun(s, r, err)
		}
		if err = a.acceptBulk(ctx, s, v, r, created); err != nil {
			return err
		}
		r.Stage = "uploading"
		if err = s.SaveRun(r); err != nil {
			return err
		}
		if err = auth.Client.Bulk.Upload(ctx, created, file, size); err != nil {
			r.Stage = "upload_uncertain"
			_ = s.SaveRun(r)
			return &Error{Code: "bulk_upload_incomplete", Message: "The accepted job ID is retained; upload was not automatically retried.", RunID: r.ID, RemoteID: r.RemoteID, Next: "Inspect bulk get; explicitly retry bulk upload with the original file while its URL remains valid.", Cause: err}
		}
		r.Stage = "uploaded"
		if err = s.SaveRun(r); err != nil {
			return err
		}
		if _, err = auth.Client.Bulk.Start(ctx, r.RemoteID); err != nil {
			return &Error{Code: "bulk_start_incomplete", Message: "Start outcome requires a safe status read; the accepted ID is retained.", RunID: r.ID, RemoteID: r.RemoteID, Cause: err}
		}
		r.Stage = "running"
		if err = s.SaveRun(r); err != nil {
			return err
		}
		delete(v.Uploads, r.RemoteID)
		delete(v.BulkRuns, r.ID)
		if err = s.Save(ctx, v); err != nil {
			return err
		}
		return a.recoverRun(ctx, s, auth, r, f.Output)
	}
	bulk.AddCommand(run)
	var createInput string
	cf := &queryFlags{}
	create := &cobra.Command{Use: "create", Short: "Create a bulk draft once and privately retain its signed upload capability", Args: cobra.NoArgs}
	bindRun(create, cf, false)
	create.Flags().StringVar(&createInput, "input", "", "private CreateBulkJobRequest JSON file, or - for stdin")
	create.RunE = func(cmd *cobra.Command, args []string) error {
		b, err := a.readInput(cmd, createInput, 256<<10)
		if err != nil {
			return err
		}
		if createInput != "-" {
			if _, err = state.ReadPrivate(createInput, 256<<10); err != nil {
				return fail("private_config_required", "Bulk configuration files must have private ownership and permissions.", "")
			}
		}
		var request fd.CreateBulkJobRequest
		if err = decodeObject(b, &request, "source", "options", "callback_url"); err != nil {
			return err
		}
		if err = fd.ValidateBulkCreate(request); err != nil {
			return err
		}
		a.rememberSecrets(request)
		if err = ensureOutput(cf.Output); err != nil {
			return err
		}
		canonical, _ := json.Marshal(request)
		id, err := a.runID(*cf)
		if err != nil {
			return err
		}
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		v, err := a.load(ctx, s)
		if err != nil {
			return err
		}
		auth, err := a.authorize(ctx, v, request.Source, cf.Recurring)
		if err != nil {
			return err
		}
		if cf.DryRun {
			return a.success(map[string]any{"dryRun": true, "request": sanitize(request)}, "", false, nil)
		}
		r, exists, err := a.begin(ctx, s, v, auth, id, "bulk:create", canonical, []fd.Source{request.Source}, cf.Recurring)
		if err != nil {
			return err
		}
		if exists {
			return a.existing(s, r, cf.Output)
		}
		created, err := auth.Client.Bulk.Create(ctx, request)
		if err != nil {
			return a.failureRun(s, r, err)
		}
		if err = a.acceptBulk(ctx, s, v, r, created); err != nil {
			return err
		}
		return a.accepted(r, cf.Output, false)
	}
	bulk.AddCommand(create)
	var uploadInput, uploadRun string
	upload := &cobra.Command{Use: "upload BULK_ID", Short: "Explicitly upload the original ID file to a saved draft", Args: cobra.ExactArgs(1)}
	upload.Flags().StringVar(&uploadInput, "input", "", "original video-ID file")
	upload.Flags().StringVar(&uploadRun, "run-id", "", "owning local run ID (auto-detected when unique)")
	upload.RunE = func(cmd *cobra.Command, args []string) error {
		file, hash, size, err := validateIDs(cmd.Context(), uploadInput)
		if err != nil {
			return err
		}
		defer closeIDs(file)
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		v, err := a.load(ctx, s)
		if err != nil {
			return err
		}
		r, err := findBulkRun(s, v, args[0], uploadRun)
		if err != nil {
			return err
		}
		auth, err := a.authorizeRun(ctx, v, r)
		if err != nil {
			return err
		}
		if err = checkRunAccount(auth, r); err != nil {
			return err
		}
		if r.InputFileHash != "" && r.InputFileHash != hash {
			return fail("bulk_input_conflict", "This job is bound to another input file.", "Use the original bytes; do not overwrite accepted input.")
		}
		created := v.Uploads[args[0]]
		if created == nil {
			return fail("upload_capability_missing", "No original upload capability is stored.", "Preserve the accepted job and consult its status; no replacement was created.")
		}
		job, err := auth.Client.Bulk.Get(ctx, args[0])
		if err != nil {
			return err
		}
		if job.Status != "awaiting_upload" {
			return fail("bulk_not_awaiting_upload", "Input cannot be replaced after the job starts.", "")
		}
		r.InputFileHash = hash
		r.Stage = "uploading"
		if err = s.SaveRun(r); err != nil {
			return err
		}
		if err = auth.Client.Bulk.Upload(ctx, created, file, size); err != nil {
			r.Stage = "upload_uncertain"
			_ = s.SaveRun(r)
			return &Error{Code: "bulk_upload_incomplete", Message: "The original job and input fingerprint are retained.", RunID: r.ID, RemoteID: args[0], Cause: err}
		}
		r.Stage = "uploaded"
		if err = s.SaveRun(r); err != nil {
			return err
		}
		return a.success(map[string]any{"id": args[0], "status": "uploaded", "nextCommand": "flashdata bulk start " + args[0]}, r.ID, false, nil)
	}
	bulk.AddCommand(upload)
	for _, action := range []string{"get", "start", "pause", "resume", "cancel", "wait", "results", "update"} {
		action := action
		var output, runID, patchInput string
		cmd := &cobra.Command{Use: action + " BULK_ID", Short: action + " an existing bulk job; never creates replacement work", Args: cobra.ExactArgs(1)}
		if action == "results" {
			cmd.Flags().StringVar(&output, "output", "", "new NDJSON receipt output file")
			cmd.Flags().StringVar(&runID, "run-id", "", "local recovery record to reuse")
		}
		if action == "update" {
			cmd.Flags().StringVar(&patchInput, "input", "", "private partial options/callback JSON file, or - for stdin")
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			if err := ensureOutput(output); err != nil {
				return err
			}
			ctx, cancel := a.context(cmd)
			defer cancel()
			s, unlock, err := a.lock(ctx)
			if err != nil {
				return err
			}
			defer unlock()
			v, err := a.load(ctx, s)
			if err != nil {
				return err
			}
			auth, err := a.authorize(ctx, v, "", false)
			if err != nil {
				return err
			}
			job, err := auth.Client.Bulk.Get(ctx, args[0])
			if err != nil {
				return err
			}
			if err = requireSource(auth, fd.Source(job.Source)); err != nil {
				return err
			}
			var result any
			switch action {
			case "get":
				result = job
			case "start":
				if job.Status != "awaiting_upload" {
					result = job
				} else {
					result, err = auth.Client.Bulk.Start(ctx, args[0])
				}
			case "pause":
				result, err = auth.Client.Bulk.Pause(ctx, args[0])
			case "resume":
				if job.Status != "paused" {
					return fail("bulk_not_paused", "The job is not paused; resume was not sent.", "")
				}
				result, err = auth.Client.Bulk.Resume(ctx, args[0])
			case "cancel":
				result, err = auth.Client.Bulk.Cancel(ctx, args[0])
			case "wait":
				result, err = auth.Client.Bulk.Wait(ctx, args[0], fd.PollOptions{Timeout: a.Timeout})
			case "update":
				b, e := a.readInput(cmd, patchInput, 256<<10)
				if e != nil {
					return e
				}
				if patchInput != "-" {
					if _, e = state.ReadPrivate(patchInput, 256<<10); e != nil {
						return fail("private_config_required", "Patch credentials must be in a private file or stdin.", "")
					}
				}
				var raw struct {
					Options  map[string]any  `json:"options"`
					Callback json.RawMessage `json:"callback_url"`
				}
				if err = decodeObject(b, &raw, "options", "callback_url"); err != nil {
					return err
				}
				patch := fd.UpdateBulkJobRequest{Options: raw.Options}
				if string(raw.Callback) == "null" {
					patch.RemoveCallback = true
				} else if len(raw.Callback) > 0 {
					var callback string
					if json.Unmarshal(raw.Callback, &callback) != nil {
						return fail("invalid_callback", "callback_url must be a string or null.", "")
					}
					patch.CallbackURL = &callback
				}
				a.rememberSecrets(patch)
				result, err = auth.Client.Bulk.Update(ctx, args[0], patch)
			case "results":
				id, e := a.runID(queryFlags{RunID: runID})
				if e != nil {
					return e
				}
				canonical, _ := json.Marshal(map[string]string{"bulkId": args[0]})
				r, exists, e := a.begin(ctx, s, v, auth, id, "bulk:results", canonical, []fd.Source{fd.Source(job.Source)}, false)
				if e != nil {
					return e
				}
				if !exists {
					r.Status = "accepted"
					r.RemoteID = args[0]
					if e = s.SaveRun(r); e != nil {
						return e
					}
				}
				return a.recoverRun(ctx, s, auth, r, output)
			}
			if err != nil {
				return &Error{Code: "bulk_operation_failed", Message: "Existing bulk work was not replaced.", RemoteID: args[0], Cause: err}
			}
			if updated, ok := result.(*fd.BulkJob); ok {
				if err = a.recordBulkState(ctx, s, v, auth, updated); err != nil {
					return &Error{Code: "bulk_local_state_not_saved", Message: "The remote operation completed; inspect the existing job before any further mutation.", RemoteID: args[0], Cause: err}
				}
			}
			return a.success(sanitize(result), "", false, nil)
		}
		bulk.AddCommand(cmd)
	}
	var cursor string
	var listLimit int
	list := &cobra.Command{Use: "list", Short: "List existing remote bulk jobs", Args: cobra.NoArgs}
	list.Flags().StringVar(&cursor, "cursor", "", "opaque page cursor")
	list.Flags().IntVar(&listLimit, "limit", 20, "page size, 1–50")
	list.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, err := a.store()
		if err != nil {
			return err
		}
		v, err := a.load(ctx, s)
		if err != nil {
			return err
		}
		auth, err := a.authorize(ctx, v, "", false)
		if err != nil {
			return err
		}
		result, err := auth.Client.Bulk.List(ctx, fd.BulkListParams{Cursor: cursor, Limit: listLimit})
		if err != nil {
			return err
		}
		return a.success(sanitize(result), "", false, nil)
	}
	bulk.AddCommand(list)
}

func (a *App) recordBulkState(ctx context.Context, s *state.Store, v *state.Vault, auth *access, job *fd.BulkJob) error {
	records, err := s.Runs()
	if err != nil {
		return err
	}
	for _, summary := range records {
		if summary.RemoteID != job.ID && v.BulkRuns[summary.ID] != job.ID {
			continue
		}
		if checkRunAccount(auth, &summary) != nil {
			continue
		}
		r, err := s.ReadRun(summary.ID)
		if err != nil {
			return err
		}
		if r == nil {
			return state.ErrCorrupt
		}
		r.RemoteID = job.ID
		r.Stage = job.Status
		if r.Status == "intent" || r.Status == "uncertain" {
			r.Status = "accepted"
		}
		if err = s.SaveRun(r); err != nil {
			return err
		}
	}
	if job.Status != "awaiting_upload" && (v.Uploads[job.ID] != nil) {
		delete(v.Uploads, job.ID)
		for id, remoteID := range v.BulkRuns {
			if remoteID == job.ID {
				delete(v.BulkRuns, id)
			}
		}
		return s.Save(ctx, v)
	}
	return nil
}
func bulkRequest(source fd.Source, optionsPath, storagePath string) (fd.CreateBulkJobRequest, error) {
	options := map[string]any{}
	if optionsPath != "" {
		b, err := state.ReadPrivate(optionsPath, 256<<10)
		if err != nil {
			return fd.CreateBulkJobRequest{}, fail("private_config_required", "Source options must be in a private readable file.", "")
		}
		if err = decodeStrict(b, &options); err != nil {
			return fd.CreateBulkJobRequest{}, err
		}
		if options == nil {
			return fd.CreateBulkJobRequest{}, fail("invalid_options", "Provide an options object.", "")
		}
	}
	if storagePath != "" {
		b, err := state.ReadPrivate(storagePath, 128<<10)
		if err != nil {
			return fd.CreateBulkJobRequest{}, fail("private_config_required", "Storage configuration must have private permissions.", "")
		}
		var storage map[string]any
		if err = decodeStrict(b, &storage); err != nil {
			return fd.CreateBulkJobRequest{}, err
		}
		field := "result_storage"
		if source == fd.SourceYouTubeDownload {
			field = "storage"
		}
		if _, exists := options[field]; exists {
			return fd.CreateBulkJobRequest{}, fail("storage_conflict", "Provide storage in only one file.", "")
		}
		options[field] = storage
	}
	request := fd.CreateBulkJobRequest{Source: source, Options: options}
	if err := fd.ValidateBulkCreate(request); err != nil {
		return request, err
	}
	return request, nil
}

var videoPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

func validateIDs(ctx context.Context, path string) (*os.File, string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", 0, fail("ids_unreadable", "Provide a readable video-ID file.", "")
	}
	bad := func(err error) (*os.File, string, int64, error) { f.Close(); return nil, "", 0, err }
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > 2<<30 {
		return bad(fail("invalid_ids_file", "Input must be a regular file of 1 byte–2 GiB.", ""))
	}
	snapshot, err := state.CreateTempOutput(os.TempDir(), "flashdata-ids-")
	if err != nil {
		return bad(fail("ids_snapshot_failed", "Could not create a private ID snapshot; no bulk job was created.", "Check temporary disk space and permissions."))
	}
	snapshotPath := snapshot.Name()
	committed := false
	defer func() {
		snapshot.Close()
		if !committed {
			os.Remove(snapshotPath)
		}
	}()
	h := sha256.New()
	reader := bufio.NewScanner(io.TeeReader(io.LimitReader(f, (2<<30)+1), io.MultiWriter(h, snapshot)))
	reader.Buffer(make([]byte, 4096), 65536)
	count := 0
	for reader.Scan() {
		if err := ctx.Err(); err != nil {
			return bad(err)
		}
		if !videoPattern.MatchString(strings.TrimSpace(reader.Text())) {
			return bad(fail("invalid_video_id", "Each line must contain one 11-character YouTube ID.", ""))
		}
		count++
	}
	if reader.Err() != nil || count == 0 {
		return bad(fail("invalid_ids_file", "The ID file is empty or unreadable.", ""))
	}
	if err = snapshot.Sync(); err != nil {
		return bad(err)
	}
	snapshotInfo, err := snapshot.Stat()
	if err != nil || snapshotInfo.Size() < 1 || snapshotInfo.Size() > 2<<30 {
		return bad(fail("invalid_ids_file", "Input changed size or exceeded the supported limit during validation.", ""))
	}
	if err = snapshot.Close(); err != nil {
		return bad(err)
	}
	f.Close()
	stable, err := os.Open(snapshotPath)
	if err != nil {
		return nil, "", 0, err
	}
	committed = true
	return stable, hex.EncodeToString(h.Sum(nil)), snapshotInfo.Size(), nil
}

func closeIDs(file *os.File) { file.Close(); _ = os.Remove(file.Name()) }
func (a *App) acceptBulk(ctx context.Context, s *state.Store, v *state.Vault, r *state.Run, created *fd.BulkJobCreated) error {
	r.Status = "accepted"
	r.Stage = "awaiting_upload"
	r.RemoteID = created.Job.ID
	r.Receipt, _ = json.Marshal(created.Job)
	if v.Uploads == nil {
		v.Uploads = map[string]*fd.BulkJobCreated{}
	}
	if v.BulkRuns == nil {
		v.BulkRuns = map[string]string{}
	}
	v.Uploads[r.RemoteID] = created
	v.BulkRuns[r.ID] = r.RemoteID
	if err := s.Save(ctx, v); err != nil {
		_ = s.SaveRun(r)
		return &Error{Code: "accepted_bulk_not_saved", Message: "Preserve the accepted bulk ID; secure upload-capability storage failed.", RunID: r.ID, RemoteID: r.RemoteID}
	}
	if err := s.SaveRun(r); err != nil {
		return &Error{Code: "accepted_bulk_not_saved", Message: "The accepted bulk is linked in the private vault; recover the existing run.", RunID: r.ID, RemoteID: r.RemoteID}
	}
	return nil
}
func findBulkRun(s *state.Store, v *state.Vault, bulkID, runID string) (*state.Run, error) {
	if runID != "" {
		r, err := s.ReadRun(runID)
		if err != nil {
			return nil, err
		}
		if r == nil || (r.RemoteID != bulkID && v.BulkRuns[runID] != bulkID) {
			return nil, fail("bulk_run_mismatch", "The selected run is not linked to that bulk job.", "")
		}
		if r.RemoteID == "" {
			r.RemoteID = bulkID
			r.Status = "accepted"
		}
		return r, nil
	}
	runs, err := s.Runs()
	if err != nil {
		return nil, err
	}
	var found *state.Run
	for i := range runs {
		r := &runs[i]
		if r.RemoteID == bulkID || v.BulkRuns[r.ID] == bulkID {
			if found != nil {
				return nil, fail("ambiguous_bulk_run", "Select the owning --run-id explicitly.", "")
			}
			found = r
		}
	}
	if found == nil {
		return nil, fail("bulk_run_not_found", "No local run owns this upload capability.", "Inspect runs list.")
	}
	full, err := s.ReadRun(found.ID)
	if err != nil {
		return nil, err
	}
	if full == nil {
		return nil, state.ErrCorrupt
	}
	found = full
	if found.RemoteID == "" {
		found.RemoteID = bulkID
		found.Status = "accepted"
	}
	return found, nil
}
