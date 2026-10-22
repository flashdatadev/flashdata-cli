# Bulk data and media delivery

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

See the complete commands in the [English README](../README.md#complete-bulk-workflow) or [简体中文](../README.zh-CN.md). Supported sources are youtube_metadata, youtube_transcript, youtube_captions, youtube_trainability and youtube_download. Options are validated through the Go SDK before job creation.

`bulk run` checks the ID file, private option/storage files, output path, scopes and recurring requirements, then commits intent. It creates one job, atomically stores the signed upload descriptor and run-to-job link in the protected vault, commits the remote ID in the journal, uploads once, starts and waits. A crash between vault and journal commits can recover the remote ID from the private link. The public receipt never includes the upload URL.

A repeated existing run never uploads or creates another task. `runs recover` performs safe status/result reads only. If input/start is missing, it returns explicit instructions for `bulk upload` and `bulk start`. Upload binds to the SHA-256 of the original file; another file conflicts. Explicit same-file PUT retry is only permitted while the same draft awaits input and the original URL remains valid. A lost start response is resolved by Get; an already-started job is not started again.

Bulk control commands operate on existing IDs: get/list/start/pause/resume/cancel/wait/results/update. `resume` requires a paused job and an explicit command after correcting its cause. `update --input` accepts a private JSON object containing partial options and callback_url (string or null to remove). Data result copies use options.result_storage; media download destination uses options.storage. Server-side credential merging and provider rules remain authoritative.

Manual metadata inputs are provided as examples/bulk-create.json and examples/bulk-update.json. Both create and update accept --input -; pipe these non-secret examples directly on PowerShell or redirect them on Bash as described in the command reference and source inputs. Apply the metadata options example before upload/start. Once a scan is complete, the backend permits only the documented credential-only result_storage patch on a paused customer-delivery job; a generic options patch is not valid in that state. A paused coordinator may need time to release its lease before updates are allowed. Control commands refresh local stages, and signed upload capabilities are discarded after the job leaves awaiting_upload.

ID files contain one 11-character video ID per line, from 1 byte through 2 GiB. Upload capability expires after 15 minutes; internal results remain for 24 hours. Customer buckets follow their own retention. SDK limits bound each compressed/decompressed part to 64 MiB and each row to 16 MiB by default. CLI result export streams across parts without assembling the whole dataset in memory.

Validation streams the IDs into a private temporary snapshot. The hash, byte count and upload all use that snapshot, so editing or appending to the original file during API calls cannot change the validated upload. Temporary disk space up to the input size is required, including for a dry-run. Snapshots are removed when the command returns; a forcibly killed process may leave a private flashdata-ids-*.tmp file in the OS temporary directory. Preserve active snapshots until the owning process exits before any manual cleanup.

S3, S3-compatible, OSS, GCS, Azure, COS, BOS and TOS shapes are supported. storage_url is a bucket/container name. GCS requires a service account; Azure needs account_name and a write-capable SAS; provider endpoint/region/permissions must be valid. No example invents real credentials. Media files are delivered into your cloud destination; JSONL reports outcomes, not video bytes.

Check each row's status, even for a completed job. The CLI preserves invalid/failed rows and exits nonzero for item failures. A malformed/interrupted stream leaves a partial export; inspect/recover by run ID using a fresh output filename. Disk failures preserve the original intent/remote ID wherever possible, and never authorize replaying creation. An expired upload capability is not renewed by creating another job automatically.

## Lost responses and partial receipts

A lost start response keeps the original bulk ID and input fingerprint. `runs recover` first reads current remote state; after observing that input upload is no longer permitted it removes spent upload capabilities, then fetches results without another create, upload or start. Partial gzip/result transfers remain partial during cached exports, retain available rows and exit nonzero. Corrected safe result reads can complete the original journal.

MP3 audio and MP4 video use the same lifecycle; choose the matching private options file
shown in the README. See the [SDK and CLI comparison](sdk-parity.md) for all source
coverage, media examples and recovery differences.
