# JSON, plain text and Markdown

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Response-format flags are available in current `main` source. Published v1.0.1
binaries do not include them yet. With Go 1.25+, build the current source using
`go install github.com/flashdatadev/flashdata-cli/cmd/flashdata@main`.

`--accept` chooses the API response type. `--respond-with markdown` selects
Markdown content. They are independent HTTP headers, outside query JSON and params.

| Flags | API response |
| --- | --- |
| Neither flag | Original JSON fields |
| `--accept text/plain` | Original fields as indented text |
| `--respond-with markdown` | JSON with Markdown in `content`, replacing `results` |
| `--accept text/plain --respond-with markdown` | Metadata followed by Markdown |

Markdown preserves source links, transcript timestamps and response metadata. The
API formats retrieved content without generating a summary.

## Search and transcripts

```bash
flashdata search "vector databases" --accept text/plain
flashdata search "vector databases" --respond-with markdown
flashdata youtube transcript dQw4w9WgXcQ --accept text/plain --respond-with markdown
```

Each realtime command submits a query. For all other sources, use `query --input`
with the same flags. `--output FILE` saves the selected response body to a new file.

Plain text goes directly to stdout. Add `--json` to keep the CLI's versioned JSON
envelope, with the text in `data`. For submission receipts, the text is in
`data.submission`, alongside saved remote IDs and recovery instructions. JSON API
responses keep the existing CLI envelope. `--output` writes the response body;
stdout then reports the output path. Credentials and signed URLs are redacted.

## Background jobs and batches

```bash
flashdata search "vector databases" --mode async --accept text/plain --run-id research
flashdata batch submit --input examples/batch.json --accept text/plain --run-id research-batch
flashdata jobs get JOB_ID --accept text/plain
flashdata jobs wait JOB_ID --accept text/plain
flashdata jobs result JOB_ID --accept text/plain --respond-with markdown
```

Submissions return receipts; status reads return status. They support `--accept`.
Choose Markdown when reading completed results. `--wait` on an async query or
batch also accepts `--respond-with markdown` for completed results; the submission
request receives no Markdown header. Batch waiting returns one result per job,
with job labels separating plain-text results.

The CLI saves accepted job IDs before waiting or writing output. A malformed
receipt leaves an uncertain run with any candidate IDs available for inspection.
Neither rerunning that ID nor recovery submits replacement work. Access checks
and polling use JSON regardless of the selected final response format.

YouTube Download supports text receipts and Markdown result manifests. Media bytes
keep their original format. Bulk exports remain JSONL; account, discovery and
onboarding commands keep their existing formats.

## Saved runs and recovery

```bash
flashdata runs recover research --accept text/plain --respond-with markdown
flashdata runs show research --result
```

The run journal remembers the saved response's format. Reusing a run ID without
format flags returns that saved format after access verification. A different
explicit format for a completed background job performs safe result reads, without
another query or query-credit charge. The same applies to `jobs result`.

A saved realtime response cannot be fetched again in another format. An explicit
format change returns `run_response_format_conflict` without repeating the query.
Use `runs show RUN_ID --result` to read the existing response. To clear a saved
Markdown choice for a background result, pass `--respond-with ""`.

The CLI accepts `application/json` and `text/plain` for `--accept`, and `markdown`
or an empty `--respond-with`. Invalid choices fail before API calls. Errors go to
stderr with a nonzero exit status; `--json` keeps structured CLI errors.
See the [API response-format guide](https://flashdata.dev/docs/guides/response-formats)
for HTTP requests and response bodies.
