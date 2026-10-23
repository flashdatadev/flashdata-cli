# Local CLI review

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

This records the local review before initial GitHub source publication. Current workflow results are available in [GitHub Actions](https://github.com/flashdatadev/flashdata-cli/actions). Source publication does not create version tags, GitHub Releases or backend deployments.

Reviewed against the supplied CLI spec, the checked-out backend implementation and the sibling Go SDK. No live credentials or production mutations were used. The final joint review also updates the sibling Go SDK; the existing backend identity patch remains undeployed.

## Response-format review

Current source adds independent response headers while retaining JSON defaults,
saved job IDs and the CLI output envelope. API-generated fixtures cover all 14
realtime sources and 15 stored-result sources, text receipts, status, waiting and
recovery. Review fixes redact signed URLs inside both Markdown response types,
sanitize usage metadata and retain every candidate ID from malformed receipts.
Format changes never repeat a paid query. See [response formats](response-formats.md).

## Initial publication review

| Area | Gap fixed |
| --- | --- |
| Environment overrides | Bulk recovery writes preserve the original saved credentials, proofs and phase instead of replacing them with an environment-only profile. |
| Key import | Importing the same saved key verifies and reuses it without deleting onboarding/upload recovery data; explicit replacement remains available. |
| Onboarding | Pending registration status gives a recovery command without creating access; whitespace in a repeated scope flag does not change the saved request. Verified connection fallback retains connection/expiry metadata. |
| Source discovery | Secret redaction no longer corrupts public credential-property schemas. Restricted keys can inspect public schemas while pricing/permission status remains scoped. |
| Batch access | One current access/identity check covers every distinct source, avoiding per-item reads against the management API rate limit. Dry-runs return available estimates by source and request count. |
| Submission recovery | Failed commits after accepted batches return remote IDs in structured errors. No replacement submission is made. |
| Recurring policy | Newly created recurring records retain their owner/installation requirement through flag omission, result reads and recovery. |
| Journal integrity | Large batch records can be read back; history lists omit full payloads. Orphaned results block replay when a journal record is missing. |
| Bulk recovery | Controls refresh local stages and remove spent upload capabilities. Uploads use a private snapshot matching the validated hash and byte count even if the original file changes. |
| Private input/output | Generic query/batch files containing credentials or callbacks require private permissions. Windows output files receive a private ACL at creation. Callback metadata and accidental argument errors do not echo secret values. |
| Bulk JSON/flags | Create/update field names are case sensitive; update accepts stdin. Realtime --wait, ignored job-output flags and runs --output without --result are rejected. |
| Migration | A single source snapshot is hashed and decoded. Interrupted imports are source-bound and cannot replace unrelated history. Completed import retries preserve the current profile, including a later logout. |
| Linux credential store | Explicit existing local Unix D-Bus endpoints only; no session-bus autolaunch or remote plaintext Secret Service transport. |
| Documentation | Matching English/Chinese commands, runnable manual bulk JSON inputs, PowerShell stdin examples and accurate state-specific control guidance. |

The optional check_backend_contracts.cjs development check compares all 15 SDK source definitions against the actual backend catalog, verifies the backend's filters:null compatibility, and exercises registration plus all five bulk-source option validators. It starts no server, database or cloud client. Default Go tests additionally submit every source-discovery example through the shared command path against a loopback API and exercise recovery/security failures.

See the [joint release-readiness check](release-readiness.md) for retained candidate IDs, partial-export errors, lost-response acceptance and deterministic allowlisted packages.

See [validation evidence](validation.md) for exact executed results. Native Linux/macOS credential-store checks, the race detector, production feature availability and backend deployment are not established by this Windows review. That local review preceded source publication; the CLI now uses an immutable public SDK dependency without a sibling replacement. No version tag or GitHub Release is created by the source push.
