# Local validation evidence

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

This records the local review before initial GitHub source publication. Current workflow results are available in [GitHub Actions](https://github.com/flashdatadev/flashdata-cli/actions). Source publication does not create version tags, GitHub Releases or backend deployments.

Validated locally on Windows amd64 on 2026-10-22, including the final joint SDK/CLI review against the CLI spec and checked-out backend. See the [review findings](review.md) and [joint release-readiness report](release-readiness.md). Tests use loopback mock APIs and dummy credentials only. No publication, deployment, real trial registration, paid API request or customer-cloud write was performed.

## Executed checks

| Check | Local result |
| --- | --- |
| CLI, Go 1.27.3: `go test -count=1 -json -coverprofile=.cache/coverage-readiness.out ./...` | 85 passing tests/subtests; zero failures/skips |
| CLI, Go 1.27.3: `go vet ./...` | Passed |
| CLI, Go 1.25.0: `go test -count=1 -json ./...` and `go vet ./...` | 85 passing tests/subtests; zero failures/skips, including compiled-binary/crash acceptance; vet passed |
| Go SDK, Go 1.27.3: `go test -count=1 -json ./...` and `go vet ./...` | Final SDK: 158 passing tests/subtests on both Go 1.27.3 and 1.25.0, including all 22 examples and both README programs; both vet runs pass |
| Python packaging/documentation scripts: Ruff check and format check | Passed |
| Source inspection | No real-key/private-key/token patterns found; test credentials are synthetic |
| Documentation | Local Markdown targets resolve; every Markdown document links to FlashData; all 18 English/Chinese README code blocks match; example JSON/query validation passed |
| Actual backend contracts | All 15 source definitions match the checked-out catalog; backend null-filter compatibility, registration and all five bulk-source option validators pass without starting services |
| Command reference | Matches the final executable's offline command tree: 52 public command nodes, including command groups |
| Standalone builds | Linux amd64/arm64, macOS amd64/arm64 and Windows amd64 build with `CGO_ENABLED=0` |
| Local archives | Five matching archives and SHA256SUMS; archive members checked against the binary/docs/example allowlist, with no state, caches, credentials or development runtimes |
| Packaged Windows executable | Extracted executable runs `--version`, `--help` and `commands --json` offline without creating profile state |

The recorded full-suite coverage is 75.0% for internal/app and 59.2% for internal/state on this Windows host. Subprocess execution is checked by acceptance assertions and is not merged into these coverage percentages; main has no in-process coverage. Platform-specific code for other operating systems is not counted here. JSON test logs and coverage files remain under the ignored .cache directories, rather than shipping inside archives.

## Behaviors checked

- Real Windows DPAPI round-trip/tamper detection and private output ACLs.
- A compiled executable registers, queries, claims, saves the candidate, launches its actual fresh verifier process, then acknowledges installation against a mock protocol-2 server.
- Registration proof and exact input exist on disk before POST; a lost response reuses the same proof. A failed fresh verifier never sends ACK.
- Repeated recurring execution across processes uses its saved result only after live access verification; revocation blocks cached reads.
- Intent is persisted before mutation; unknown/failed outcomes cannot automatically resubmit. Input/account conflicts and missing/corrupt credential state fail closed.
- Kernel process locks exclude another writer and are released after the lock holder is killed.
- Query normalization, duplicate-field rejection, source permissions, numeric/provenance retention, secret redaction, dry-run behavior and stdin deadlines.
- Async and mixed-source batch acceptance/recovery; batch --wait resumes accepted work without a second submission.
- Bulk create/upload/start/wait, manual controls, original input-file binding, streamed JSONL receipts, per-item failure retention and current-access checks before reading saved results.
- Existing batch/bulk output files are rejected before submission and are never overwritten.
- Explicit migration preserves the original legacy file and blocks replay of original run IDs. A source-bound marker protects interrupted migration; repeated completed migration preserves current credentials/logout.
- Environment-key bulk updates preserve saved key/proof/phase. Same-key import and explicit import into an environment-only profile retain accepted-job recovery.
- Recurring policy survives flag omission and result recovery. Orphaned results prevent another paid submission.
- Large journals round-trip, history lists omit full payloads, and accepted-batch commit errors expose remote IDs for recovery.
- Public schema credential definitions remain intact with strict top-level fields and background callback metadata. Every one of the 15 source-discovery examples passes through the CLI query path.
- Private generic credential-bearing input files, bulk patch stdin/case-sensitive fields, manual bulk state updates and spent upload-capability cleanup.
- Validated bulk ID snapshots remain unchanged when the original source file is edited. Snapshots are private from creation and cleaned on command return.

## Validation still required before a public release

Native Linux/macOS execution and actual Secret Service/Keychain interactions have not been run from this Windows host. Cross-compilation does not prove native credential-store behavior. The prepared CI uses private-file fixtures on Unix; real OS-store validation remains a separate platform check. The race detector was not run locally because a supported C compiler is unavailable; a Linux race job is prepared but has not run on GitHub.

The initial local review used a sibling SDK replacement. Source publication replaces it with an immutable public pseudo-version and checksum; CI checks out only the CLI repository and resolves the SDK through go.mod. The local review did not run remote CI. Current checks are linked above; source publication does not create a GitHub Release, sign binaries or notarize macOS executables.

Imported-key recurring execution needs the accompanying API-key identity backend patch deployed; unsupported identity endpoints block recurring use. Claimed connections can use their existing authoritative connection identity endpoint. Onboarding also depends on the server's feature gates. This task does not deploy or enable backend features, and it does not verify production credentials, credits or customer bucket permissions.

## Final joint package and recovery evidence

- Full Go 1.27.3 and 1.25.0 suites passed with no skipped acceptance tests. Reports use `.cache/test-results-readiness-127.jsonl` and `.cache/test-results-readiness-125.jsonl`.
- Compiled-process acceptance covers lost ACK recovery with the same candidate/proof and lost bulk start recovery without repeated create/upload/start; spent upload capabilities are removed after reconciliation.
- Malformed SDK response fields retain candidate IDs through CLI errors and durable uncertain records. Partial exports and per-item failures return nonzero status with error/export metadata on stderr and no success envelope.
- Both repositories pass four archive-policy unit tests (including unsafe-path subcases); reviewed file manifests, deterministic ZIP metadata and copied dependency notices are included. Python lint/format and backend source/bulk/registration contract checks pass.
- All 18 public FlashData website, docs, pricing and source-documentation links returned HTTP 200 on credential-free HEAD checks. This does not exercise paid APIs or account state.
- Final packages are checked against their manifests and SHA-256 checksums. The SDK verifier installs a clean packaged consumer through a file-only proxy and cross-builds all five targets. CLI reports record both runtime source hashes and the build toolchain.

## Public macOS CI coverage

After source publication, the native Intel macOS Go 1.25 and 1.27 tests and vet checks passed. Apple Silicon jobs using `macos-latest`, `macos-26` and `macos-15` never received a hosted runner, so no native Apple Silicon test result is claimed. The workflow now uses `macos-15-intel` for full native macOS tests and cross-builds the Apple Silicon executable with CGo disabled on both supported Go versions. The archive job still builds all five distribution targets. Current results are in [GitHub Actions](https://github.com/flashdatadev/flashdata-cli/actions).

This verifies compilation of the Apple Silicon binary; native Apple Silicon runtime and live Keychain integration still require separate acceptance checks. Superseded CI runs cancel automatically. No release, version tag or backend deployment is part of these checks.
