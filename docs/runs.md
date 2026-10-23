# Durable runs and duplicate protection

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

A run belongs to one profile, account/credential identity and pair of service origins. Files use SHA-256-derived names rather than caller-provided paths. Run IDs allow 1–160 letters/digits plus dot, underscore, colon, at, plus and hyphen; the first character must be a letter/digit. Normalize the query before hashing: flat and nested parameters have the same wire form, object key order does not matter, and conflicting/duplicate fields are rejected. Source and mode are part of execution intent. Output path, waiting and recurring flags do not create a second execution identity.

Before any paid mutation, a private, synced and atomically replaced intent record exists. Its request metadata is sanitized; cloud credentials stay in private configuration/vaults. Same ID/different input or account/origin conflicts. Same successful run returns its saved result after live access/source checks. A missing/corrupt result is an error, not permission to submit again. Credentials revoked since the first execution stop cached reads.

New recurring records also retain their owner/installation requirement. Omitting --recurring on a later invocation or using runs show/recover cannot weaken that saved policy. An orphaned result whose journal record is missing blocks another submission with the same ID. Journal records have a 64 MiB serialized limit, checked before writing; this accommodates the CLI's 8 MiB batch input limit and JSON escaping. runs list returns metadata without full query inputs or acceptance receipts; runs show inspects an individual complete record.

| Local status | Meaning and recovery |
| --- | --- |
| intent | Process may have crashed before or after dispatch. No automatic replay. |
| uncertain | Response was lost; inspect account/server records before an intentional new execution. |
| failed | A known rejection/local failure occurred; the ID is retained and is not automatically retried. |
| accepted | Known job/bulk ID is durable. `runs recover` reads existing remote work. |
| partial | A bulk result stream stopped; partial JSONL is kept and safe reads may be repeated. |
| succeeded | Result hash and file were committed; verify current access before cache/export. |

Realtime uncertain requests have no server receipt lookup in the current API. There is deliberately no force-retry switch that silently converts uncertain history into new paid work. One-off commands generate an ID; repeating without it is a new execution.

Async/batch acceptance receipts retain remote IDs and usage. Recovery waits/fetches using those IDs only. Batch aggregation is bounded to 128 MiB; use individual `jobs result --output` for larger results. Bulk rows stream to a separate private file with a running checksum; partially completed exports are retained before an error is reported. Cache/export verifies the stored checksum. Failed/invalid item receipts cause nonzero exit after they have been saved.

Batch access metadata is verified once per command and checked against every distinct source, avoiding per-item identity reads. Batch dry-runs include available source-specific estimates. If an accepted batch cannot be committed locally, its remote IDs are returned in error.details.remoteIds for explicit recovery. Bulk get/start/control refreshes local stages and removes spent upload capabilities after the job leaves awaiting_upload.

The profile's kernel lock covers command mutation, including network submission and final state commit. Concurrent writers wait only within the command deadline. The lock file is not removed; Windows LockFileEx and POSIX flock release when the process exits/crashes. This avoids unsafe stale-lock unlink/reclaim races. A fresh installation child only reads the atomically committed candidate, without acquiring the writer lock.

Run history has no automatic expiry or eviction. Profiles retain history on logout. Large results are separate from the credential vault. The host owns cron, Task Scheduler, Codex or other recurring invocation; the CLI writes no scheduler configuration and stores no key in scheduler text. This is local duplicate protection, not cross-machine exactly-once execution.

## Response formats

The journal retains the selected response format. Cached results reuse that format
when no flags are supplied. A different explicit format never causes another paid
query: realtime cache changes fail with `run_response_format_conflict`; completed
background results can be read again using `runs recover --accept ... --respond-with ...`.
Plain-text receipts still retain confirmed job IDs, and malformed receipts retain
candidate IDs on an uncertain run. See [response formats](response-formats.md).

## Candidate IDs and partial exports

Malformed responses can leave `candidateRemoteIds` on an uncertain record. These IDs do not populate the confirmed `remoteId`/`remoteIds` fields. Inspect each candidate with the original account and reconcile the entire request; repeating the run or `runs recover` does not promote candidate evidence or resubmit work.

A saved partial NDJSON export always returns a `partial_result` error, even if every row received so far completed. Available rows can still be written to a new output file. Per-item failure exports return `bulk_item_failures`; failure metadata goes to stderr with no success envelope on stdout.
