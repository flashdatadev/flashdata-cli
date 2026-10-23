package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
)

type queryFlags struct {
	responseFlags
	Input, RunID, Mode, Output string
	DryRun, Recurring, Wait    bool
}

func bindRun(cmd *cobra.Command, f *queryFlags, input bool) {
	if input {
		cmd.Flags().StringVar(&f.Input, "input", "", "JSON file, or - for stdin")
	}
	cmd.Flags().StringVar(&f.RunID, "run-id", "", "stable execution ID; required for recurring work")
	cmd.Flags().StringVar(&f.Output, "output", "", "new result or submission-receipt output file")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "validate without a paid submission")
	cmd.Flags().BoolVar(&f.Recurring, "recurring", false, "require verified owned/installed access")
}
func bindQuery(cmd *cobra.Command, f *queryFlags, input bool) {
	bindRun(cmd, f, input)
	bindResponse(cmd, &f.responseFlags, true)
	cmd.Flags().StringVar(&f.Mode, "mode", "realtime", "realtime or async")
	cmd.Flags().BoolVar(&f.Wait, "wait", false, "wait for an accepted async job")
}
func (a *App) addQueries() {
	f := &queryFlags{}
	q := &cobra.Command{Use: "query", Short: "Validate and execute a Data API JSON request", Args: cobra.NoArgs}
	bindQuery(q, f, true)
	q.RunE = func(cmd *cobra.Command, args []string) error {
		b, err := a.readInput(cmd, f.Input, 1<<20)
		if err != nil {
			return err
		}
		return a.query(cmd, b, *f)
	}
	a.Root.AddCommand(q)
	search := &cobra.Command{Use: "search QUERY", Short: "Google Search with the normal query billing and recovery path", Args: cobra.ExactArgs(1)}
	sf := &queryFlags{}
	bindQuery(search, sf, false)
	var num, page int64
	var gl, hl string
	search.Flags().Int64Var(&num, "num", 0, "requested number of results")
	search.Flags().Int64Var(&page, "page", 0, "requested page; separately billed")
	search.Flags().StringVar(&gl, "gl", "", "country code")
	search.Flags().StringVar(&hl, "hl", "", "interface language")
	search.RunE = func(cmd *cobra.Command, args []string) error {
		p := map[string]any{}
		for _, v := range []struct {
			name  string
			value any
		}{{"num", num}, {"page", page}, {"gl", gl}, {"hl", hl}} {
			if cmd.Flags().Changed(v.name) {
				p[v.name] = v.value
			}
		}
		b, _ := json.Marshal(map[string]any{"source": "google_search", "query": args[0], "params": p})
		return a.query(cmd, b, *sf)
	}
	a.Root.AddCommand(search)
	yt := &cobra.Command{Use: "youtube", Short: "YouTube search, metadata, existing captions and media delivery"}
	a.Root.AddCommand(yt)
	for _, action := range []string{"search", "metadata", "transcript", "captions"} {
		action := action
		flags := &queryFlags{}
		var language string
		var limit int64
		cmd := &cobra.Command{Use: action + " INPUT", Short: "YouTube " + action + " through the shared query workflow", Args: cobra.ExactArgs(1)}
		bindQuery(cmd, flags, false)
		if action == "transcript" || action == "captions" {
			cmd.Flags().StringVar(&language, "language", "en", "caption language; existing tracks only")
		}
		if action == "search" {
			cmd.Flags().Int64Var(&limit, "max-results", 0, "requested number of search results")
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			body := map[string]any{"source": "youtube_" + action}
			params := map[string]any{}
			if action == "search" {
				body["query"] = args[0]
				if cmd.Flags().Changed("max-results") {
					params["max_results"] = limit
				}
			} else {
				body["video_id"] = args[0]
			}
			if action == "transcript" || action == "captions" {
				params["language"] = language
				params["auto"] = action == "transcript"
			}
			body["params"] = params
			b, _ := json.Marshal(body)
			return a.query(cmd, b, *flags)
		}
		yt.AddCommand(cmd)
	}
	var storagePath, video string
	df := &queryFlags{Mode: "async"}
	download := &cobra.Command{Use: "download VIDEO_ID", Short: "Deliver media to private customer storage (background only)", Args: cobra.ExactArgs(1)}
	bindQuery(download, df, false)
	_ = download.Flags().Set("mode", "async")
	download.Flags().StringVar(&storagePath, "storage-config", "", "private customer-storage JSON file")
	var outputFormat string
	download.Flags().StringVar(&outputFormat, "format", "mp3", "output format")
	download.RunE = func(cmd *cobra.Command, args []string) error {
		video = args[0]
		b, err := state.ReadPrivate(storagePath, 128<<10)
		if err != nil {
			return fail("storage_config_unreadable", "Customer storage must be a private, readable JSON file.", "")
		}
		var storage map[string]any
		if err = decodeStrict(b, &storage); err != nil {
			return err
		}
		raw, _ := json.Marshal(map[string]any{"source": "youtube_download", "video_id": video, "params": map[string]any{"storage": storage, "download_type": "audio", "output_format": outputFormat}})
		return a.query(cmd, raw, *df)
	}
	yt.AddCommand(download)
	a.addJobs()
	a.addBatch()
}

// normalizeQuery accepts the spec's flat parameters and the canonical SDK params
// object, rejecting duplicates, mixed collisions and unsupported fields.
func normalizeQuery(b []byte, async bool) (fd.Query, []byte, error) {
	var raw map[string]any
	if err := decodeStrict(b, &raw); err != nil {
		return fd.Query{}, nil, err
	}
	if raw == nil {
		return fd.Query{}, nil, fail("invalid_query", "Expected a query object.", "")
	}
	source, ok := raw["source"].(string)
	if !ok {
		return fd.Query{}, nil, fail("invalid_source", "Provide a source identifier.", "")
	}
	schema, err := fd.QuerySchema(fd.Source(source))
	if err != nil {
		return fd.Query{}, nil, err
	}
	var definition struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	_ = json.Unmarshal(schema, &definition)
	var pSchema struct {
		Properties map[string]any `json:"properties"`
	}
	_ = json.Unmarshal(definition.Properties["params"], &pSchema)
	params := map[string]any{}
	if value, exists := raw["params"]; exists {
		var valid bool
		params, valid = value.(map[string]any)
		if !valid {
			return fd.Query{}, nil, fail("invalid_params", "params must be a JSON object.", "")
		}
	}
	for key, value := range raw {
		if key == "callback_url" || definition.Properties[key] != nil {
			continue
		}
		if _, ok := pSchema.Properties[key]; !ok {
			return fd.Query{}, nil, fail("unsupported_field", "Unsupported query field: "+key, "")
		}
		if _, exists := params[key]; exists {
			return fd.Query{}, nil, fail("parameter_conflict", "Parameter is present both at the top level and inside params: "+key, "")
		}
		params[key] = value
		delete(raw, key)
	}
	if len(params) > 0 {
		raw["params"] = params
	} else {
		delete(raw, "params")
	}
	normalized, _ := json.Marshal(raw)
	var query fd.Query
	if err = decodeStrict(normalized, &query); err != nil {
		return query, nil, err
	}
	if err = fd.ValidateQuery(query, async); err != nil {
		return query, nil, err
	}
	canonical, _ := json.Marshal(query)
	return query, canonical, nil
}
func sanitize(value any) any {
	b, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var v any
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.UseNumber()
	_ = decoder.Decode(&v)
	var visit func(any) any
	visit = func(v any) any {
		switch v := v.(type) {
		case map[string]any:
			for key, x := range v {
				normalized := strings.ToLower(strings.ReplaceAll(key, "_", ""))
				switch normalized {
				case "apikey", "accesskey", "secretkey", "privatekey", "sastoken", "authorization", "setupcredential", "handoffproof", "registrationproof", "callbackurl":
					v[key] = "[REDACTED]"
				default:
					v[key] = visit(x)
				}
			}
		case []any:
			for i, x := range v {
				v[i] = visit(x)
			}
		}
		return v
	}
	return visit(v)
}
func (a *App) runID(f queryFlags) (string, error) {
	if f.Recurring && f.RunID == "" {
		return "", fail("stable_run_id_required", "Recurring runs need an explicit --run-id.", "")
	}
	id := f.RunID
	if id == "" {
		nonce, err := fd.NewRegistrationProof()
		if err != nil {
			return "", err
		}
		id = "run-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce[:12]
	}
	if !state.ValidRunID(id) {
		return "", fail("invalid_run_id", "Use 1–160 letters, digits, dot, underscore, colon, at, plus or hyphen, starting with a letter or digit.", "")
	}
	return id, nil
}
func (a *App) begin(ctx context.Context, s *state.Store, v *state.Vault, auth *access, id, kind string, input []byte, sources []fd.Source, recurring bool) (*state.Run, bool, error) {
	hash := state.Digest(append([]byte(kind+"\x00"), input...))
	r, err := s.ReadRun(id)
	if err != nil {
		return nil, false, err
	}
	account := auth.AccountRef + "@" + state.Digest([]byte(authOrigin(auth)))
	if r != nil {
		if r.Recurring && !auth.Identity.RecurringAllowed {
			return nil, false, &Error{Code: "owned_access_required", Message: "This recurring run still requires verified installed owner access.", RunID: id}
		}
		if r.InputHash != hash || r.Kind != kind {
			return nil, false, &Error{Code: "run_id_input_conflict", Message: "This run ID already belongs to different input.", RunID: id, Next: "Inspect runs show; choose a new ID only for an intentional new execution."}
		}
		if r.AccountRef != account {
			return nil, false, &Error{Code: "run_account_mismatch", Message: "This run belongs to another account or origin.", RunID: id, Next: "Use the original profile access; no resubmission was made."}
		}
		return r, true, nil
	}
	legacy, err := s.LegacyConflict(id)
	if err != nil {
		return nil, false, err
	}
	if legacy {
		return nil, false, &Error{Code: "legacy_run_requires_inspection", Message: "This stable ID exists in migrated legacy history; it cannot be resubmitted automatically.", RunID: id, Next: "Inspect runs list/show for legacy records before intentionally choosing a new execution ID."}
	}
	// Callers hold the profile lock. Initialize before journaling intent so the
	// next command can load pinned origins; Save omits the environment API key.
	if v.Phase == "environment" && !s.Exists() {
		if err = s.Save(ctx, v); err != nil {
			return nil, false, err
		}
	}
	r = &state.Run{ID: id, Kind: kind, InputHash: hash, AccountRef: account, Status: "intent", Sources: sources, Recurring: recurring}
	if len(sources) == 1 {
		r.Source = sources[0]
	}
	r.Input, _ = json.Marshal(sanitize(json.RawMessage(input)))
	if err = s.SaveRun(r); err != nil {
		return nil, false, err
	}
	return r, false, nil
}

// origin binding is supplied through the profile, independent of key changes.
func authOrigin(auth *access) string { return auth.Origin }
func (a *App) failureRun(s *state.Store, r *state.Run, err error) error {
	r.Status = "failed"
	if fd.IsSubmissionUnknown(err) {
		r.Status = "uncertain"
		var unknown *fd.SubmissionUnknownError
		if errors.As(err, &unknown) && len(unknown.JobIDs) > 0 {
			r.CandidateRemoteIDs = append([]string(nil), unknown.JobIDs...)
		}
	}
	r.ErrorCode = a.classify(err).Code
	if saveErr := s.SaveRun(r); saveErr != nil {
		return &Error{Code: "run_record_write_failed", Message: "The persisted intent must be treated as uncertain.", RunID: r.ID, RemoteID: r.RemoteID, Next: "Preserve the run and reconcile it; do not resubmit automatically.", Cause: err}
	}
	return &Error{Code: r.ErrorCode, Message: a.classify(err).Message, RunID: r.ID, RemoteID: r.RemoteID, Next: "Inspect runs show " + r.ID + "; failed or uncertain work is not automatically retried.", Cause: err, Details: candidateDetails(r)}
}

func candidateDetails(r *state.Run) any {
	if len(r.CandidateRemoteIDs) == 0 {
		return nil
	}
	return map[string]any{"candidateRemoteIds": r.CandidateRemoteIDs, "acceptanceConfirmed": false}
}
func (a *App) existing(s *state.Store, r *state.Run, output string) error {
	if r.Status == "succeeded" {
		if r.ResultFormat == "ndjson" {
			return a.exportRun(s, r, output, true)
		}
		b, err := s.ReadResult(r)
		if err != nil {
			return err
		}
		return a.emitResult(r, b, output, true)
	}
	if r.Status == "accepted" && len(r.Receipt) > 0 {
		return a.accepted(r, output, true)
	}
	return &Error{Code: "run_requires_recovery", Message: "The run is " + r.Status + "; no request was resubmitted.", RunID: r.ID, RemoteID: r.RemoteID, Next: "Inspect runs show " + r.ID + ". Reconcile candidate IDs with jobs get or bulk get; do not replay the whole submission.", Details: candidateDetails(r)}
}
func (a *App) emitResult(r *state.Run, b []byte, output string, cached bool) error {
	// Earlier CLI versions could cache Markdown before recognizing escaped
	// credentials. Sanitize again at export, including file outputs and batches.
	if r.ResultFormat != "ndjson" {
		response := &fd.FormattedResponse{JSON: b}
		if r.ResultFormat == "text" {
			response.JSON, response.Text = nil, string(b)
		}
		var err error
		b, _, err = a.responseBytes(response)
		if err != nil {
			return err
		}
	}
	if output != "" && output != "-" {
		f, err := state.CreateOutput(output)
		if err != nil {
			if r.ID == "" {
				return &Error{Code: "output_write_failed", Message: "Could not create result output. Choose a new output path."}
			}
			return &Error{Code: "output_write_failed", Message: "The result is saved in the local run journal; choose another output path.", RunID: r.ID}
		}
		_, err = f.Write(b)
		if err == nil {
			err = f.Sync()
		}
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			if r.ID == "" {
				return &Error{Code: "output_write_failed", Message: "A partial export may remain. Read the job result again to a new output path."}
			}
			return &Error{Code: "output_write_failed", Message: "A partial export is retained; recover from the saved run.", RunID: r.ID}
		}
		return a.success(map[string]any{"output": output, "bytes": len(b), "itemFailures": r.ItemFailures}, r.ID, cached, r.Usage)
	}
	if r.ResultFormat == "ndjson" {
		return a.success(map[string]any{"resultFileAvailable": true, "rowsFormat": "ndjson", "itemFailures": r.ItemFailures, "nextCommand": "flashdata runs show " + r.ID + " --result --output FILE"}, r.ID, cached, r.Usage)
	}
	if r.ResultFormat == "text" {
		return a.writeText(string(b), r.ID, cached, r.Usage)
	}
	return a.success(json.RawMessage(b), r.ID, cached, r.Usage)
}
func (a *App) query(cmd *cobra.Command, b []byte, f queryFlags) error {
	if f.Mode != "realtime" && f.Mode != "async" {
		return fail("invalid_mode", "mode must be realtime or async.", "")
	}
	if f.Wait && f.Mode != "async" {
		return fail("wait_requires_async", "--wait requires --mode async.", "")
	}
	if err := f.responseFlags.validate(f.Mode == "realtime" || f.Wait); err != nil {
		return err
	}
	q, canonical, err := normalizeQuery(b, f.Mode == "async")
	if err != nil {
		return err
	}
	a.rememberSecrets(q)
	if err = requirePrivateInput(f.Input, q); err != nil {
		return err
	}
	id, err := a.runID(f)
	if err != nil {
		return err
	}
	if err = ensureOutput(f.Output); err != nil {
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
	auth, err := a.authorize(ctx, v, q.Source, f.Recurring)
	if err != nil {
		return err
	}
	if f.DryRun {
		pricing, e := auth.Client.Account.Pricing(ctx, fd.PricingParams{Source: q.Source, RequestCount: 1})
		return a.success(map[string]any{"dryRun": true, "request": sanitize(q), "mode": f.Mode, "estimate": pricing, "estimateAvailable": e == nil, "note": "Estimate only; admission and actual usage are determined by FlashData."}, "", false, nil)
	}
	r, exists, err := a.begin(ctx, s, v, auth, id, "query:"+f.Mode, canonical, []fd.Source{q.Source}, f.Recurring)
	if err != nil {
		return err
	}
	if exists {
		format := selectedResponse(cmd, r, f.responseFlags)
		if f.Wait && r.Status == "accepted" {
			return a.recoverRun(ctx, s, auth, r, f.Output, format)
		}
		if format != savedResponse(r) {
			if r.Status == "succeeded" && r.Kind == "query:async" {
				return a.recoverRun(ctx, s, auth, r, f.Output, format)
			}
			return formatConflict(r)
		}
		return a.existing(s, r, f.Output)
	}
	recordResponse(r, f.responseFlags)
	if err := s.SaveRun(r); err != nil {
		return err
	}
	if f.Mode == "async" {
		response, err := auth.Client.Queries.SubmitWithFormat(ctx, q, f.responseFlags.envelope())
		if err != nil {
			return a.failureRun(s, r, err)
		}
		if err := a.receiptResponse(r, response, []fd.Query{q}, false); err != nil {
			return a.failureRun(s, r, err)
		}
		r.Status = "accepted"
		if err = s.SaveRun(r); err != nil {
			return &Error{Code: "accepted_run_not_saved", Message: "The server accepted a job; preserve the remote ID.", RunID: id, RemoteID: r.RemoteID, Next: "Use jobs get/wait/result with this remote ID; do not resubmit."}
		}
		if f.Wait {
			return a.recoverRun(ctx, s, auth, r, f.Output)
		}
		return a.accepted(r, f.Output, false)
	}
	result, err := auth.Client.Queries.RealtimeWithFormat(ctx, q, f.responseFlags.content())
	if err != nil {
		return a.failureRun(s, r, err)
	}
	data, err := a.saveResponse(s, r, result, f.responseFlags)
	if err != nil {
		return &Error{Code: "result_commit_failed", Message: "The query completed but durable result commit failed; the intent must not be replayed.", RunID: id, Next: "Inspect the run and preserve any existing result files."}
	}
	return a.emitResult(r, data, f.Output, false)
}
func (a *App) addJobs() {
	jobs := &cobra.Command{Use: "jobs", Short: "Read, wait for and recover accepted background jobs"}
	a.Root.AddCommand(jobs)
	for _, action := range []string{"get", "wait", "result"} {
		action := action
		var output string
		format := &responseFlags{}
		cmd := &cobra.Command{Use: action + " JOB_ID", Short: action + " an accepted job; never submits work", Args: cobra.ExactArgs(1)}
		bindResponse(cmd, format, action == "result")
		if action == "result" {
			cmd.Flags().StringVar(&output, "output", "", "new result output file")
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			if err := format.validate(action == "result"); err != nil {
				return err
			}
			if err := ensureOutput(output); err != nil {
				return err
			}
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
			job, err := auth.Client.Jobs.Get(ctx, args[0])
			if err != nil {
				return err
			}
			if err = requireSource(auth, fd.Source(job.Source)); err != nil {
				return err
			}
			if action == "get" {
				if format.Accept == "text/plain" {
					response, err := auth.Client.Jobs.GetWithFormat(ctx, args[0], format.envelope())
					if err != nil {
						return err
					}
					return a.emitResponse(response, output)
				}
				return a.success(sanitize(job), "", false, nil)
			}
			if action == "wait" {
				result, err := auth.Client.Jobs.Wait(ctx, args[0], fd.PollOptions{Timeout: a.Timeout})
				if err != nil {
					return err
				}
				if format.Accept == "text/plain" {
					response, err := auth.Client.Jobs.GetWithFormat(ctx, args[0], format.envelope())
					if err != nil {
						return err
					}
					return a.emitResponse(response, output)
				}
				return a.success(sanitize(result), "", false, nil)
			}
			data, err := auth.Client.Jobs.ResultWithFormat(ctx, args[0], format.content())
			if err != nil {
				return err
			}
			return a.emitResponse(data, output)
		}
		jobs.AddCommand(cmd)
	}
}
func (a *App) addBatch() {
	batch := &cobra.Command{Use: "batch", Short: "Background mixed-source batch submission"}
	a.Root.AddCommand(batch)
	f := &queryFlags{}
	submit := &cobra.Command{Use: "submit", Short: "Validate all input and submit a batch once", Args: cobra.NoArgs}
	bindRun(submit, f, true)
	bindResponse(submit, &f.responseFlags, true)
	submit.Flags().BoolVar(&f.Wait, "wait", false, "wait for all accepted batch jobs")
	submit.RunE = func(cmd *cobra.Command, args []string) error {
		if err := f.responseFlags.validate(f.Wait); err != nil {
			return err
		}
		b, err := a.readInput(cmd, f.Input, 8<<20)
		if err != nil {
			return err
		}
		var items []json.RawMessage
		if err = decodeStrict(b, &items); err != nil || len(items) == 0 {
			return fail("invalid_batch", "Provide a non-empty JSON array of source queries.", "")
		}
		queries := []fd.Query{}
		sources := []fd.Source{}
		seenSources := map[fd.Source]bool{}
		for _, item := range items {
			q, _, err := normalizeQuery(item, true)
			if err != nil {
				return err
			}
			queries = append(queries, q)
			if !seenSources[q.Source] {
				sources = append(sources, q.Source)
				seenSources[q.Source] = true
			}
		}
		a.rememberSecrets(queries)
		if err = requirePrivateInput(f.Input, queries); err != nil {
			return err
		}
		if err = ensureOutput(f.Output); err != nil {
			return err
		}
		canonical, _ := json.Marshal(queries)
		id, err := a.runID(*f)
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
		auth, err := a.authorizeSources(ctx, v, sources, f.Recurring)
		if err != nil {
			return err
		}
		if f.DryRun {
			estimates := map[fd.Source]any{}
			for _, source := range sources {
				count := 0
				for _, query := range queries {
					if query.Source == source {
						count++
					}
				}
				price, e := auth.Client.Account.Pricing(ctx, fd.PricingParams{Source: source, RequestCount: count})
				estimates[source] = map[string]any{"requestCount": count, "available": e == nil, "pricing": price}
			}
			return a.success(map[string]any{"dryRun": true, "queries": sanitize(queries), "count": len(queries), "estimates": estimates, "note": "Estimates only; download cost depends on delivered bytes, and server admission is authoritative."}, "", false, nil)
		}
		r, exists, err := a.begin(ctx, s, v, auth, id, "batch", canonical, sources, f.Recurring)
		if err != nil {
			return err
		}
		if exists {
			format := selectedResponse(cmd, r, f.responseFlags)
			if f.Wait && r.Status == "accepted" {
				return a.recoverRun(ctx, s, auth, r, f.Output, format)
			}
			if format != savedResponse(r) {
				if r.Status == "succeeded" {
					return a.recoverRun(ctx, s, auth, r, f.Output, format)
				}
				return formatConflict(r)
			}
			return a.existing(s, r, f.Output)
		}
		recordResponse(r, f.responseFlags)
		if err := s.SaveRun(r); err != nil {
			return err
		}
		result, err := auth.Client.Queries.BatchWithFormat(ctx, queries, f.responseFlags.envelope())
		if err != nil {
			return a.failureRun(s, r, err)
		}
		r.Status = "accepted"
		if err := a.receiptResponse(r, result, queries, true); err != nil {
			return a.failureRun(s, r, err)
		}
		if err = s.SaveRun(r); err != nil {
			return &Error{Code: "accepted_run_not_saved", Message: "Batch accepted; preserve these remote IDs.", RunID: id, Next: "Use jobs get/wait/result with these IDs; do not submit a replacement batch.", Details: map[string]any{"remoteIds": r.RemoteIDs, "usage": r.Usage}}
		}
		if f.Wait {
			return a.recoverRun(ctx, s, auth, r, f.Output)
		}
		return a.accepted(r, f.Output, false)
	}
	batch.AddCommand(submit)
}
