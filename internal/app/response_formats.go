package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/flashdatadev/flashdata-cli/internal/state"
	fd "github.com/flashdatadev/flashdata-go"
	"github.com/spf13/cobra"
)

type responseFlags struct{ Accept, RespondWith string }

func bindResponse(cmd *cobra.Command, f *responseFlags, content bool) {
	cmd.Flags().StringVar(&f.Accept, "accept", "application/json", "API response type: application/json or text/plain")
	if content {
		cmd.Flags().StringVar(&f.RespondWith, "respond-with", "", "markdown for realtime or completed results; omit for original fields")
	}
}

func (f responseFlags) validate(content bool) error {
	if f.Accept != "application/json" && f.Accept != "text/plain" {
		return fail("invalid_accept", "--accept must be application/json or text/plain.", "")
	}
	if f.RespondWith != "" && f.RespondWith != "markdown" {
		return fail("invalid_respond_with", "--respond-with must be markdown or omitted.", "")
	}
	if f.RespondWith != "" && !content {
		return fail("markdown_requires_content", "Markdown applies to realtime queries or completed results.", "Use --wait for background work, or jobs result --respond-with markdown.")
	}
	return nil
}
func (f responseFlags) content() fd.ContentOptions {
	return fd.ContentOptions{Accept: fd.Accept(f.Accept), RespondWith: fd.Representation(f.RespondWith)}
}
func (f responseFlags) envelope() fd.ResponseOptions {
	return fd.ResponseOptions{Accept: fd.Accept(f.Accept)}
}
func savedResponse(r *state.Run) responseFlags {
	f := responseFlags{Accept: r.ResponseAccept, RespondWith: r.RespondWith}
	if f.Accept == "" {
		f.Accept = "application/json"
	}
	return f
}
func selectedResponse(cmd *cobra.Command, r *state.Run, requested responseFlags) responseFlags {
	f := savedResponse(r)
	if cmd.Flags().Changed("accept") {
		f.Accept = requested.Accept
	}
	if cmd.Flags().Changed("respond-with") {
		f.RespondWith = requested.RespondWith
	}
	return f
}
func recordResponse(r *state.Run, f responseFlags) {
	r.ResponseAccept = f.Accept
	r.RespondWith = f.RespondWith
}

// Text is API content, not a JSON document. Apply the same credential and signed
// URL redaction used by structured output without removing provenance queries.
var responseURL = regexp.MustCompile(`https?://[^\s<>"']+`)
var responseSecretField = regexp.MustCompile(`(?im)^(\s*(?:\*\*)?(?:api key|access key|secret key|private key|sas token|authorization|setup credential|handoff proof|registration proof|callback url):(?:\*\*)?)[^\n]*`)
var markdownSecretPattern = regexp.MustCompile(`(?i)\b(?:fd\\_|fda\\_setup\\_|fda\\_handoff\\_)(?:[A-Za-z0-9]|\\[_-])+`)

// Match the API's escaped values without unescaping or otherwise rewriting the
// returned document. Newlines in scalar values use Markdown hard line breaks.
func markdownSecret(secret string) string {
	var escaped strings.Builder
	for _, char := range secret {
		switch char {
		case '&':
			escaped.WriteString("&amp;")
		case '<':
			escaped.WriteString("&lt;")
		case '>':
			escaped.WriteString("&gt;")
		default:
			if strings.ContainsRune("\\`*_{}[]()#+!|.-", char) {
				escaped.WriteByte('\\')
			}
			escaped.WriteRune(char)
		}
	}
	text := strings.ReplaceAll(escaped.String(), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.ReplaceAll(text, "\n", "  \n")
}

func (a *App) redactedResponseText(text string) string {
	for _, secret := range a.knownSecrets {
		if secret != "" {
			text = strings.ReplaceAll(text, markdownSecret(secret), "[REDACTED]")
			text = strings.ReplaceAll(text, secret, "[REDACTED]")
		}
	}
	text = secretPattern.ReplaceAllString(text, "[REDACTED]")
	text = markdownSecretPattern.ReplaceAllString(text, "[REDACTED]")
	text = responseURL.ReplaceAllStringFunc(text, func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil {
			return raw
		}
		for key := range u.Query() {
			switch strings.ToLower(key) {
			case "x-amz-signature", "x-goog-signature", "sig", "signature", "sas_token":
				return "[REDACTED signed URL]"
			}
		}
		return raw
	})
	return responseSecretField.ReplaceAllString(text, `${1} "[REDACTED]"`)
}
func (a *App) responseBytes(v *fd.FormattedResponse) ([]byte, string, error) {
	if v.JSON == nil {
		return []byte(a.redactedResponseText(v.Text)), "text", nil
	}
	value := a.redacted(v.JSON)
	b, err := json.Marshal(value)
	return b, "", err
}
func responseUsage(v *fd.FormattedResponse) json.RawMessage {
	if v.JSON == nil {
		inUsage := false
		for _, line := range strings.Split(v.Text, "\n") {
			if strings.HasPrefix(line, "Usage:") {
				value := strings.TrimSpace(strings.TrimPrefix(line, "Usage:"))
				if strings.HasPrefix(value, "{") && json.Valid([]byte(value)) {
					return json.RawMessage(value)
				}
				inUsage = value == ""
				continue
			}
			if line != "" && !strings.HasPrefix(line, " ") {
				inUsage = false
			}
			if inUsage && strings.HasPrefix(line, "  Credits:") {
				value := strings.TrimSpace(strings.TrimPrefix(line, "  Credits:"))
				var number json.Number
				if json.Unmarshal([]byte(value), &number) == nil && number.String() != "" && len(value) > 0 && value[0] != '"' {
					return json.RawMessage(`{"credits":` + value + `}`)
				}
			}
		}
		return nil
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(v.JSON, &fields)
	return fields["usage"]
}
func (a *App) responseUsage(v *fd.FormattedResponse) json.RawMessage {
	raw := responseUsage(v)
	if raw == nil {
		return nil
	}
	b, _ := json.Marshal(a.redacted(raw))
	return b
}
func (a *App) saveResponse(s *state.Store, r *state.Run, v *fd.FormattedResponse, f responseFlags) ([]byte, error) {
	b, format, err := a.responseBytes(v)
	if err != nil {
		return nil, err
	}
	next := *r
	next.ResultFormat = format
	recordResponse(&next, f)
	if usage := a.responseUsage(v); usage != nil {
		next.Usage = usage
	}
	if err := s.SaveResult(&next, b); err != nil {
		return nil, err
	}
	*r = next
	return b, nil
}
func (a *App) emitResponse(v *fd.FormattedResponse, output string) error {
	b, format, err := a.responseBytes(v)
	if err != nil {
		return err
	}
	return a.emitResult(&state.Run{ResultFormat: format}, b, output, false)
}
func formatConflict(r *state.Run) error {
	return &Error{Code: "run_response_format_conflict", Message: "This run already has a saved response in another format; no query was repeated.", RunID: r.ID, Next: "Read it with runs show " + r.ID + " --result. For background jobs, use jobs result or runs recover with the desired response flags."}
}

var receiptJobID = regexp.MustCompile(`^job_[A-Za-z0-9_-]+$`)

// The API's documented text receipts contain JSON-quoted scalar fields with
// fixed indentation. Extract only identity fields, validate every receipt against
// submitted sources, and retain candidates on malformed acceptance. Never infer
// acceptance from a URL, arbitrary prose, or a partial set of batch IDs.
func textReceipts(text string, queries []fd.Query, batch bool) ([]fd.AsyncSubmission, error) {
	rows := []map[string]json.RawMessage{}
	current := -1
	if !batch {
		rows = append(rows, map[string]json.RawMessage{})
		current = 0
	}
	valid := true
	inQueries := false
	candidates := []string{}
	candidateSet := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		depth := len(line) - len(trimmed)
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			if trimmed != "" {
				valid = false
			}
			continue
		}
		value = strings.TrimSpace(value)
		if key == "Id" && ((!batch && depth == 0) || (batch && inQueries && depth == 4)) {
			var candidate string
			if json.Unmarshal([]byte(value), &candidate) == nil && receiptJobID.MatchString(candidate) && !candidateSet[candidate] {
				candidates = append(candidates, candidate)
				candidateSet[candidate] = true
			}
		}
		if batch {
			if depth == 0 {
				inQueries = key == "Queries" && value == ""
				continue
			}
			if !inQueries {
				continue
			}
			if depth == 2 {
				index, err := strconv.Atoi(key)
				if err != nil || index != len(rows)+1 || value != "" {
					valid = false
					current = -1
					continue
				}
				rows = append(rows, map[string]json.RawMessage{})
				current = len(rows) - 1
				continue
			}
			if depth != 4 {
				continue
			}
		} else if depth != 0 {
			continue
		}
		if current < 0 {
			continue
		}
		if key != "Id" && key != "Source" && key != "Status" {
			continue
		}
		if _, exists := rows[current][key]; exists {
			valid = false
			continue
		}
		rows[current][key] = json.RawMessage(value)
	}
	receipts := make([]fd.AsyncSubmission, 0, len(rows))
	seen := map[string]bool{}
	for i, row := range rows {
		var receipt fd.AsyncSubmission
		for key, out := range map[string]*string{"Id": &receipt.ID, "Source": &receipt.Source, "Status": &receipt.Status} {
			if err := json.Unmarshal(row[key], out); err != nil {
				valid = false
			}
		}
		if !receiptJobID.MatchString(receipt.ID) || seen[receipt.ID] {
			valid = false
		}
		seen[receipt.ID] = true
		if i >= len(queries) || receipt.Source != string(queries[i].Source) || receipt.Status == "" {
			valid = false
		}
		receipts = append(receipts, receipt)
	}
	if !valid || len(rows) != len(queries) {
		return nil, &fd.SubmissionUnknownError{Cause: &fd.ProtocolError{Message: "inconsistent text submission receipt"}, JobIDs: candidates}
	}
	return receipts, nil
}
func (a *App) receiptResponse(r *state.Run, v *fd.FormattedResponse, queries []fd.Query, batch bool) error {
	r.Usage = a.responseUsage(v)
	if v.JSON == nil {
		receipts, err := textReceipts(v.Text, queries, batch)
		if err != nil {
			return err
		}
		r.Receipt, _ = json.Marshal(a.redactedResponseText(v.Text))
		r.ReceiptAccept = "text/plain"
		if batch {
			for _, receipt := range receipts {
				r.RemoteIDs = append(r.RemoteIDs, receipt.ID)
			}
		} else {
			r.RemoteID = receipts[0].ID
		}
		return nil
	}
	r.Receipt, _ = json.Marshal(a.redacted(v.JSON))
	r.ReceiptAccept = "application/json"
	r.Usage = a.responseUsage(v)
	if batch {
		var result fd.BatchResponse
		if err := v.DecodeJSON(&result); err != nil {
			return err
		}
		for _, receipt := range result.Queries {
			r.RemoteIDs = append(r.RemoteIDs, receipt.ID)
		}
	} else {
		var receipt fd.AsyncSubmission
		if err := v.DecodeJSON(&receipt); err != nil {
			return err
		}
		r.RemoteID = receipt.ID
	}
	return nil
}

func (a *App) writeText(text string, run string, cached bool, usage json.RawMessage) error {
	text = a.redactedResponseText(text)
	if a.JSON {
		return a.success(text, run, cached, usage)
	}
	_, err := io.WriteString(a.IO.Out, text)
	return err
}

func batchText(id string, text string) string { return fmt.Sprintf("Job: %s\n\n%s\n", id, text) }
