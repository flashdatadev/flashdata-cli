package app

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
)

func (a *App) addRuns() {
	runs := &cobra.Command{Use: "runs", Short: "Inspect durable local history and recover known accepted work"}
	a.Root.AddCommand(runs)
	var limit int
	list := &cobra.Command{Use: "list", Short: "List local execution metadata without network calls", Args: cobra.NoArgs}
	list.Flags().IntVar(&limit, "limit", 100, "maximum records displayed; history is never evicted")
	list.RunE = func(cmd *cobra.Command, args []string) error {
		if limit < 1 || limit > 10000 {
			return fail("invalid_limit", "limit must be 1–10000.", "")
		}
		s, err := a.store()
		if err != nil {
			return err
		}
		records, err := s.Runs()
		if err != nil {
			return err
		}
		if len(records) > limit {
			records = records[:limit]
		}
		return a.success(records, "", false, nil)
	}
	runs.AddCommand(list)
	var result bool
	var output string
	show := &cobra.Command{Use: "show RUN_ID", Short: "Inspect a record; --result revalidates live access before reading saved data", Args: cobra.ExactArgs(1)}
	show.Flags().BoolVar(&result, "result", false, "read the saved result after access verification")
	show.Flags().StringVar(&output, "output", "", "new result output file")
	show.RunE = func(cmd *cobra.Command, args []string) error {
		if output != "" && !result {
			return fail("result_flag_required", "--output requires --result.", "")
		}
		if err := ensureOutput(output); err != nil {
			return err
		}
		s, err := a.store()
		if err != nil {
			return err
		}
		r, err := s.ReadRun(args[0])
		if err != nil {
			return err
		}
		if r == nil {
			return fail("run_not_found", "No such local run exists.", "")
		}
		if !result {
			return a.success(r, r.ID, false, nil)
		}
		ctx, cancel := a.context(cmd)
		defer cancel()
		v, err := a.load(ctx, s)
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
		return a.exportRun(s, r, output, true)
	}
	runs.AddCommand(show)
	var recoverOutput string
	recoverFormat := &responseFlags{}
	recover := &cobra.Command{Use: "recover RUN_ID", Short: "Read accepted remote jobs and save results; never resubmit a query or create a job", Args: cobra.ExactArgs(1)}
	bindResponse(recover, recoverFormat, true)
	recover.Flags().StringVar(&recoverOutput, "output", "", "new result output file")
	recover.RunE = func(cmd *cobra.Command, args []string) error {
		if err := recoverFormat.validate(true); err != nil {
			return err
		}
		if err := ensureOutput(recoverOutput); err != nil {
			return err
		}
		ctx, cancel := a.context(cmd)
		defer cancel()
		s, unlock, err := a.lock(ctx)
		if err != nil {
			return err
		}
		defer unlock()
		r, err := s.ReadRun(args[0])
		if err != nil {
			return err
		}
		if r == nil {
			return fail("run_not_found", "No such local run exists.", "")
		}
		if strings.HasPrefix(r.Kind, "bulk:") && (cmd.Flags().Changed("accept") || cmd.Flags().Changed("respond-with")) {
			return fail("bulk_response_format_unsupported", "Bulk exports keep their JSONL format.", "Use --output to save the export.")
		}
		v, err := a.load(ctx, s)
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
		if r.RemoteID == "" && v.BulkRuns[r.ID] != "" {
			r.RemoteID = v.BulkRuns[r.ID]
			r.Status = "accepted"
			r.Stage = "awaiting_upload"
			if err = s.SaveRun(r); err != nil {
				return err
			}
		}
		return a.recoverRun(ctx, s, auth, r, recoverOutput, selectedResponse(cmd, r, *recoverFormat))
	}
	runs.AddCommand(recover)
}
func (a *App) authorizeRun(ctx context.Context, v *state.Vault, r *state.Run) (*access, error) {
	sources := r.Sources
	if len(sources) == 0 {
		sources = []fd.Source{r.Source}
	}
	return a.authorizeSources(ctx, v, sources, r.Recurring)
}
func checkRunAccount(auth *access, r *state.Run) error {
	if r.AccountRef != auth.AccountRef+"@"+state.Digest([]byte(authOrigin(auth))) {
		return &Error{Code: "run_account_mismatch", Message: "This saved run belongs to another account or origin.", RunID: r.ID, Next: "Restore the original profile access."}
	}
	return nil
}
func (a *App) exportRun(s *state.Store, r *state.Run, output string, cached bool) error {
	if r.ResultHash == "" {
		return fail("result_unavailable", "The run has no saved result.", "Inspect its status and remote ID.")
	}
	if r.ResultFormat != "ndjson" {
		b, err := s.ReadResult(r)
		if err != nil {
			return err
		}
		return a.emitResult(r, b, output, cached)
	}
	f, err := s.OpenResult(r)
	if err != nil {
		return err
	}
	defer f.Close()
	if output != "" && output != "-" {
		out, err := state.CreateOutput(output)
		if err != nil {
			return fail("output_write_failed", "Choose a new writable result output file.", "")
		}
		_, err = io.Copy(out, f)
		if err == nil {
			err = out.Sync()
		}
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			return &Error{Code: "output_write_failed", Message: "A partial export is retained; the journal copy is available.", RunID: r.ID}
		}
	}
	data := map[string]any{"format": "ndjson", "itemFailures": r.ItemFailures, "status": r.Status}
	if output != "" && output != "-" {
		data["output"] = output
	} else {
		data["nextCommand"] = "flashdata runs show " + r.ID + " --result --output FILE"
	}
	if r.Status == "partial" {
		return &Error{Code: "partial_result", Message: "The saved export is incomplete; available receipts have been preserved.", RunID: r.ID, Next: "Use runs recover to repeat safe result reads; do not submit replacement work.", Details: data}
	}
	if r.ItemFailures > 0 {
		return &Error{Code: "bulk_item_failures", Message: "Failed or invalid item receipts have been preserved.", RunID: r.ID, Next: "Inspect each row before deciding on a new task.", Details: data}
	}
	return a.success(data, r.ID, cached, r.Usage)
}
func (a *App) recoverRun(ctx context.Context, s *state.Store, auth *access, r *state.Run, output string, formats ...responseFlags) error {
	format := savedResponse(r)
	if len(formats) > 0 {
		format = formats[0]
	}
	if r.Status == "succeeded" && format == savedResponse(r) {
		return a.exportRun(s, r, output, true)
	}
	if r.Kind == "query:realtime" && format != savedResponse(r) {
		return formatConflict(r)
	}
	if strings.HasPrefix(r.Kind, "bulk:") {
		if r.RemoteID == "" {
			return a.existing(s, r, output)
		}
		job, err := auth.Client.Bulk.Get(ctx, r.RemoteID)
		if err != nil {
			return err
		}
		if job.Status == "awaiting_upload" {
			return &Error{Code: "bulk_input_required", Message: "This accepted job still awaits its original input/start.", RunID: r.ID, RemoteID: r.RemoteID, Next: "Use bulk upload with the original file, then bulk start; no upload or create was replayed."}
		}
		// A lost start response can leave a spent signed upload in the vault.
		// Reconcile that capability after the safe state read, before exporting.
		v, err := s.Load(ctx)
		if err != nil {
			return err
		}
		if v != nil {
			if err = a.recordBulkState(ctx, s, v, auth, job); err != nil {
				return err
			}
		}
		r.Stage = job.Status
		if _, err = auth.Client.Bulk.Wait(ctx, r.RemoteID, fd.PollOptions{Timeout: a.Timeout}); err != nil {
			return &Error{Code: "bulk_wait_stopped", Message: "Local waiting stopped; keep the accepted bulk ID.", RunID: r.ID, RemoteID: r.RemoteID, Cause: err}
		}
		r.ItemFailures = 0
		err = s.SaveStream(r, func(writer io.Writer) error {
			encoder := json.NewEncoder(writer)
			for row, err := range auth.Client.Bulk.Rows(ctx, r.RemoteID) {
				if err != nil {
					return err
				}
				if err = encoder.Encode(a.redacted(row)); err != nil {
					return err
				}
				if string(row["status"]) != "\"completed\"" {
					r.ItemFailures++
				}
			}
			return nil
		})
		if err != nil {
			if output != "" {
				_ = a.exportRun(s, r, output, false)
			}
			return &Error{Code: "partial_result", Message: "Partial result receipts were retained; recovery can repeat safe result reads.", RunID: r.ID, RemoteID: r.RemoteID, Cause: err}
		}
		return a.exportRun(s, r, output, false)
	}
	if r.Kind == "query:async" {
		if r.RemoteID == "" {
			return a.existing(s, r, output)
		}
		if _, err := auth.Client.Jobs.Wait(ctx, r.RemoteID, fd.PollOptions{Timeout: a.Timeout}); err != nil {
			return &Error{Code: "job_wait_stopped", Message: "Keep the accepted job ID and recover later.", RunID: r.ID, RemoteID: r.RemoteID, Cause: err}
		}
		data, err := auth.Client.Jobs.ResultWithFormat(ctx, r.RemoteID, format.content())
		if err != nil {
			return err
		}
		raw, err := a.saveResponse(s, r, data, format)
		if err != nil {
			return err
		}
		return a.emitResult(r, raw, output, false)
	}
	if r.Kind == "batch" {
		if len(r.RemoteIDs) == 0 {
			return a.existing(s, r, output)
		}
		results := []json.RawMessage{}
		var text strings.Builder
		size := 0
		for _, id := range r.RemoteIDs {
			if _, err := auth.Client.Jobs.Wait(ctx, id, fd.PollOptions{Timeout: a.Timeout}); err != nil {
				return &Error{Code: "batch_wait_stopped", Message: "The original batch IDs are retained.", RunID: r.ID, RemoteID: id, Cause: err}
			}
			data, err := auth.Client.Jobs.ResultWithFormat(ctx, id, format.content())
			if err != nil {
				return err
			}
			b, _, err := a.responseBytes(data)
			if err != nil {
				return err
			}
			part := string(b)
			if format.Accept == "text/plain" {
				part = batchText(id, part)
			}
			size += len(part) + 1
			if size > 128<<20 {
				return fail("result_aggregation_limit", "Batch result aggregation exceeds 128 MiB.", "Fetch individual jobs with jobs result --output FILE.")
			}
			if format.Accept == "text/plain" {
				text.WriteString(part)
			} else {
				results = append(results, b)
			}
		}
		b, _ := json.Marshal(results)
		r.ResultFormat = ""
		if format.Accept == "text/plain" {
			b = []byte(text.String())
			r.ResultFormat = "text"
		}
		recordResponse(r, format)
		if err := s.SaveResult(r, b); err != nil {
			return err
		}
		return a.emitResult(r, b, output, false)
	}
	return a.existing(s, r, output)
}
