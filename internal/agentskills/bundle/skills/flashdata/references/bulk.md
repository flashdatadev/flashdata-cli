# Bulk video datasets and media

Use bulk for a supplied video-ID list with one source and shared options. Choose individual query commands for one video, and `batch submit --input FILE` for a JSON list of mixed-source queries. MCP is optional; these bulk workflows use the standalone CLI.

| Task | Bulk source | Shared options |
| --- | --- | --- |
| Metadata | `youtube_metadata` | `include_formats`, `include_chapters` |
| Existing automatic transcripts | `youtube_transcript` | `language`, `auto: true` |
| Existing uploader captions | `youtube_captions` | `language`, `auto: false` |
| License classifications | `youtube_trainability` | No options needed for the basic request |
| Audio/video delivery | `youtube_download` | `download_type`, `output_format`, optional quality/clip settings and storage |

Use one bulk job for a supplied list of video IDs with the same source and options. Work within the user's requested scope and budget. Check `flashdata auth status`, `flashdata sources list` and `flashdata account balance` for usable access and Credits; installing a skill does not grant source permissions.

Create `ids.txt` with one 11-character video ID per line, not watch URLs, CSV or a JSON array. For example:

```text
dQw4w9WgXcQ
jNQXAC9IVRw
```

Use absolute paths in commands. The paths below are examples to replace with your working directory. Keep one `--run-id` for this intended job and reuse it during recovery. Use a different ID for a genuinely new job.

## Export metadata

```sh
flashdata bulk run --source youtube_metadata --input /absolute/work/ids.txt --run-id metadata-001 --timeout 1h --output /absolute/work/metadata.jsonl
```

Change `--source` for another data source and choose a matching run ID and output filename. To set shared options, add `--options /absolute/work/options.json`. The file contains only the options object, for example `{"language":"en","auto":true}` for transcripts; do not copy a whole single-query request with `source`, `video_id` or `params`. Use the current [bulk options reference](https://flashdata.dev/docs/guides/bulk) for supported fields. Missing caption tracks remain missing, and license labels do not establish reuse permission.

## Deliver audio or video to cloud storage

Downloads require approved owner access with `youtube_download` permission, sufficient Credits and the user's chosen writable destination. Create a private `storage.json` using the [CLI storage example](https://github.com/flashdatadev/flashdata-cli#complete-bulk-workflow). Do not put storage credentials in chat or source control.

Create `audio-options.json` with:

```json
{"download_type":"audio","output_format":"mp3"}
```

All files passed to `--options` or `--storage-config` must be private. On Linux/macOS use `chmod 600 audio-options.json storage.json`; on Windows restrict their ACLs to your user, SYSTEM and optionally Administrators.

```sh
flashdata bulk run --source youtube_download --input /absolute/work/ids.txt --storage-config /absolute/work/storage.json --options /absolute/work/audio-options.json --run-id audio-001 --timeout 1h --output /absolute/work/audio-receipts.jsonl
```

For MP4 video, use an options file containing `{"download_type":"audio_video","video_quality":"<=720","output_format":"mp4"}` instead, with a distinct run ID and output path. Choose the variant requested by the user. Media is delivered to the user's bucket/container; JSONL contains delivery receipts. For data sources, optional `--storage-config` selects customer storage for result copies instead of media.

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

For the audio example, substitute `audio-001` when inspecting or recovering its run.
