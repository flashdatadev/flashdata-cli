# Bulk video metadata

Use one bulk job for a supplied list of video IDs with the same source and options. Work within the user's requested scope and budget. Check `flashdata auth status`, `flashdata sources list` and `flashdata account balance` for usable access and Credits; installing a skill does not grant source permissions.

Create `ids.txt` with one 11-character video ID per line, not watch URLs, CSV or a JSON array. For example:

```text
dQw4w9WgXcQ
jNQXAC9IVRw
```

Use absolute paths in commands. The paths below are examples to replace with your working directory. Keep one `--run-id` for this intended job and reuse it during recovery. Use a different ID for a genuinely new job.

Copy [bulk-options.json](../assets/bulk-options.json) to a private `options.json` in your working directory and edit it for the task. It is the shared options object, without a `source`, `video_id` or `params` wrapper. All videos use these options. On Linux/macOS run `chmod 600 options.json`; on Windows restrict its ACL to your user, SYSTEM and optionally Administrators.

```sh
flashdata bulk run --source youtube_metadata --input /absolute/work/ids.txt --options /absolute/work/options.json --run-id metadata-001 --timeout 1h --output /absolute/work/metadata.jsonl
```

`bulk run` creates the job, uploads the IDs, starts processing, waits and saves the export. Add `--dry-run` to the planned command to validate inputs and access and read pricing without creating a job. This is not a guaranteed total-cost quote; download cost depends on delivered bytes.

## Read results and recover

Bulk exports stay **JSONL**, with one JSON object per line. Preserve each row's video ID, index, status, data and errors. Bulk commands do not accept `--accept` or `--respond-with`; do not append the single-query text/Markdown flags. `--json` controls the command's status envelope, not the exported file.

A completed job can contain failed or invalid items. Inspect every row; the CLI preserves those rows and exits nonzero for item failures. A partial transfer can also leave a partial file and a nonzero exit. Keep it for inspection, then recover into a fresh filename:

```sh
flashdata runs recover metadata-001 --timeout 1h --output /absolute/work/recovered.jsonl
```

Keep the same profile and state directory. A timeout stops local waiting; accepted server work continues. Recovery reads the existing job without creating, uploading or starting replacement work. If the original upload or start is incomplete, inspect the saved run and follow the CLI's explicit `bulk upload`/`bulk start` guidance using the original job and input. Do not generate a new run ID to retry an uncertain submission.

## Manage the existing job

Use the remote bulk ID reported by the CLI, which is different from your local run ID. These are separate operations to use when needed, not a sequence to execute automatically:

| Task | Command |
| --- | --- |
| Find the saved remote ID | `flashdata runs show metadata-001` |
| List bulk jobs | `flashdata bulk list` |
| Inspect status | `flashdata bulk get BULK_ID` |
| Wait for completion | `flashdata bulk wait BULK_ID --timeout 1h` |
| Export existing results | `flashdata bulk results BULK_ID --output /absolute/work/results-copy.jsonl` |
| Pause processing | `flashdata bulk pause BULK_ID` |
| Resume a paused job after fixing its cause | `flashdata bulk resume BULK_ID` |
| Cancel the job | `flashdata bulk cancel BULK_ID` |

For manual create/upload/start and supported option updates, see the [CLI bulk guide](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/bulk.md). Use `bulk` commands for bulk IDs; `jobs get` and `jobs result` are for individual async query jobs.
