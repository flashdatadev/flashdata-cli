package app

import (
	"encoding/json"
	"github.com/flashdatadev/flashdata-cli/internal/state"
)

func (a *App) accepted(r *state.Run, output string, cached bool) error {
	data := map[string]any{"status": "accepted", "stage": r.Stage, "submission": r.Receipt, "remoteId": r.RemoteID, "remoteIds": r.RemoteIDs, "nextCommand": "flashdata runs recover " + r.ID}
	if r.Stage == "awaiting_upload" {
		data["nextCommand"] = "flashdata bulk upload " + r.RemoteID + " --input ids.txt --run-id " + r.ID
	}
	if output != "" && output != "-" {
		file, err := state.CreateOutput(output)
		if err != nil {
			return &Error{Code: "output_write_failed", Message: "The accepted IDs are in the run journal; choose a new output file.", RunID: r.ID, RemoteID: r.RemoteID}
		}
		err = json.NewEncoder(file).Encode(a.redacted(data))
		if err == nil {
			err = file.Sync()
		}
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			return &Error{Code: "output_write_failed", Message: "The accepted IDs remain in the journal.", RunID: r.ID, RemoteID: r.RemoteID}
		}
		data["output"] = output
	}
	return a.success(data, r.ID, cached, r.Usage)
}
