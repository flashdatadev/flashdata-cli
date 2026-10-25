# Bulk data and media delivery

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Process a video-ID list with one source and shared options. `bulk run` creates the job, uploads your IDs, starts processing, waits and exports results. Use individual query commands for one video, or `batch submit` for a JSON list of mixed-source queries.

| Task | `--source` | Example shared options |
| --- | --- | --- |
| Video metadata | `youtube_metadata` | `{"include_formats":false,"include_chapters":true}` |
| Existing automatic transcripts | `youtube_transcript` | `{"language":"en","auto":true}` |
| Existing uploader captions | `youtube_captions` | `{"language":"en","auto":false}` |
| License classifications | `youtube_trainability` | `{}` |
| Audio/video delivery | `youtube_download` | `{"download_type":"audio","output_format":"mp3"}` |

Check `flashdata auth status`, `flashdata sources list` and `flashdata account balance` for your access, permitted sources and remaining Credits. Downloads require owner access and your own writable storage destination. Installing skills does not grant additional permissions. Recurring automation also requires verified owner access; see [onboarding](onboarding.md).

## Export a video dataset

Create `ids.txt` with one 11-character video ID per line:

```text
dQw4w9WgXcQ
jNQXAC9IVRw
```

Run a metadata job:

```bash
flashdata bulk run --source youtube_metadata --input ids.txt --run-id metadata-001 --timeout 1h --output metadata.jsonl
```

For transcripts, captions or license classifications, change the source, run ID and output filename. To apply shared options, save the options object in a private JSON file and add `--options options.json`. Do not include a `source`, `video_id` or `params` wrapper. All videos in the job use those options. See the [bulk API guide](https://flashdata.dev/docs/guides/bulk) for supported fields.

On Linux/macOS, protect option and storage files with `chmod 600`; on Windows, restrict their ACLs to your user, SYSTEM and optionally Administrators. The [metadata options example](../examples/bulk-options.json) can be copied into a private working directory. See [private-file requirements](security.md).

Add `--dry-run` to the planned command to validate inputs and access and read pricing without creating a job. It does not guarantee a total-cost quote. Use a stable `--run-id` for one intended job; reserve new IDs for new work.

## Deliver audio or video

Downloads require owner access and a writable cloud destination. Use the `ids.txt` created above.

### MP3 audio

Use the same `ids.txt` to request MP3 audio for multiple videos. Media files are delivered **to your bucket**; the local JSONL file contains delivery receipts.

**1. Configure your destination.** In a private working directory outside source control, create `storage.json` and replace every placeholder with your own S3 bucket and credentials:

```json
{
  "storage_type": "s3",
  "storage_url": "YOUR_BUCKET_NAME",
  "region": "us-east-1",
  "access_key": "YOUR_STORAGE_ACCESS_KEY",
  "secret_key": "YOUR_STORAGE_SECRET_KEY",
  "object_prefix": "flashdata/audio"
}
```

Create `audio-options.json` in the same directory:

```json
{"download_type":"audio","output_format":"mp3"}
```

On Linux/macOS, make both files private before running the command:

```bash
chmod 600 storage.json audio-options.json
```

On Windows, restrict both files' ACLs to your user, SYSTEM and optionally Administrators; see [private-file requirements](security.md). Other supported destinations include S3-compatible storage, OSS, GCS, Azure, COS, BOS and TOS; see [storage options](#supported-destinations).

**2. Run the complete task.** Run from the directory containing those files and `ids.txt`:

```bash
flashdata bulk run --source youtube_download --input ids.txt --storage-config storage.json --options audio-options.json --run-id audio-bulk-001 --timeout 1h --output audio-receipts.jsonl
```

The command creates, uploads, starts, waits and exports. Check `audio-receipts.jsonl` for each video's delivery status and destination. A completed job can still contain failed items; those receipts are saved and the command exits with a nonzero status.

**For one video**, use the same storage configuration:

```bash
flashdata youtube download dQw4w9WgXcQ --storage-config storage.json --format mp3 --wait --timeout 30m --output audio-delivery.json
```

### MP4 video

Use the same `ids.txt` and private `storage.json`, and create `video-options.json`:

```json
{"download_type":"audio_video","video_quality":"<=720","output_format":"mp4"}
```

On Linux/macOS, restrict the options file first:

```bash
chmod 600 video-options.json
```

```bash
flashdata bulk run --source youtube_download --input ids.txt --storage-config storage.json --options video-options.json --run-id video-bulk-001 --timeout 1h --output video-receipts.jsonl
```

Video files go to the bucket; per-video delivery outcomes go to `video-receipts.jsonl`.
On Windows, apply the same private ACL requirements as for the audio options file.

### Supported destinations

S3, S3-compatible storage, OSS, GCS, Azure, COS, BOS and TOS are supported. `storage_url` is the bucket/container name. Supply the provider's required endpoint, region and credentials; GCS needs a service account and Azure needs an account name and write-capable SAS. Keep credentials out of chat and source control. Your bucket's retention and access settings apply.

For downloads, `--storage-config` sets the media destination (`options.storage`). For data sources, it optionally requests result copies in customer storage (`options.result_storage`). Supply a destination in only one place: the storage file or the options file.

## JSONL results and failed items

Bulk exports remain **JSONL**: one JSON object per line, preserving each video's index, ID, status, data and errors. `--json` controls the command's status envelope, not the export. `bulk run` and `bulk results` do not accept `--accept` or `--respond-with`; omit those flags when recovering a bulk run too.

A completed job may still contain failed or invalid items. Inspect each row. The CLI preserves those rows and exits nonzero when items fail. Missing transcript/caption tracks are not generated or translated, and license classifications do not establish reuse permission.

An interrupted or malformed result transfer can leave a partial export. Preserve it for inspection and use a new output filename when recovering. Recovery retrieves existing results; it does not rerun failed videos.

## Monitor and control a job

The remote bulk ID identifies the server job. Your local run ID, such as `metadata-001`, identifies the CLI's saved execution. Use the appropriate ID for each command:

| Task | Command |
| --- | --- |
| Find the remote ID for a saved run | `flashdata runs show metadata-001` |
| List existing jobs | `flashdata bulk list` |
| Inspect a job | `flashdata bulk get BULK_ID` |
| Wait for completion | `flashdata bulk wait BULK_ID --timeout 1h` |
| Export existing results | `flashdata bulk results BULK_ID --output results-copy.jsonl` |
| Pause processing | `flashdata bulk pause BULK_ID` |
| Resume a paused job after fixing its cause | `flashdata bulk resume BULK_ID` |
| Cancel a job | `flashdata bulk cancel BULK_ID` |

These are separate operations to use when needed. Use `bulk` commands for bulk IDs; `jobs get` and `jobs result` handle individual async query jobs.

## Recover interrupted work

Keep the same profile and state directory, then recover into a fresh file:

```bash
flashdata runs recover metadata-001 --timeout 1h --output recovered-metadata.jsonl
```

Substitute `audio-bulk-001` or `video-bulk-001` for the download examples. A timeout stops local waiting; accepted server work continues. Recovery reads the existing job without creating, uploading or starting replacement work. If a submission outcome is unknown and no remote ID is available, inspect `runs show` and follow its guidance; do not create a new run ID to repeat the submission.

If the original upload or start is incomplete, recovery tells you which explicit step is needed. Use `bulk upload BULK_ID --input ids.txt` only with the original file while the same draft still accepts input and its upload URL is valid. Then use `bulk start BULK_ID` if needed. A lost start response is checked against the existing job before another start is attempted. An expired upload URL does not trigger a replacement job.

## Manual creation and updates

For separate steps, use `bulk create --input FILE`, `bulk upload` and `bulk start`. Copy [bulk-create.json](../examples/bulk-create.json) into a private file and edit it before creating the job. Create/update also accept `--input -` for JSON on stdin. Keep the created job ID and original ID file for upload and recovery.

`bulk update BULK_ID --input FILE` accepts partial `options` and `callback_url`; use `null` to remove the callback. See [bulk-update.json](../examples/bulk-update.json). Apply ordinary option changes before upload/start. After scanning, a paused customer-delivery job only accepts the documented credential-only `result_storage` update; it may take time to become ready for that update. Fix the cause, then explicitly resume the paused job.

## Large files and retention

ID files are UTF-8 text from 1 byte through 2 GiB. Validation needs temporary disk space up to the input size, including during a dry run. The CLI uploads a validated snapshot, so later edits to the original file do not change the submitted IDs. Normal completion removes the snapshot; a forcibly terminated process may leave a private `flashdata-ids-*.tmp` file in the OS temporary directory. Remove it only after the owning process has exited.

Upload URLs expire after 15 minutes; internal results remain for 24 hours. Save results promptly. Customer storage follows its own retention. Exports stream across result parts; the SDK defaults limit each compressed/decompressed part to 64 MiB and each row to 16 MiB.
