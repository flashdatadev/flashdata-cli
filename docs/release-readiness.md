# Joint release-readiness check

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

This records the local review before initial GitHub source publication. Current workflow results are available in [GitHub Actions](https://github.com/flashdatadev/flashdata-cli/actions). Source publication does not create version tags, GitHub Releases or backend deployments.

Reviewed together on 2026-10-22 as local 1.0.0 candidates. This check does not publish, tag, push, deploy, register real accounts or submit live paid work. Exact executed counts and platform evidence are in [validation](validation.md).

## Integration fixes

The SDK retains valid recovery IDs even when another response field has the wrong JSON type. The CLI preserves them as `candidateRemoteIds` in both errors and durable uncertain records, with `acceptanceConfirmed:false`. Candidate IDs are evidence for manual `jobs get` or `bulk get` reconciliation; they cannot turn a partial batch into a successful complete submission. Reusing the run ID never replays the submission.

Partial bulk transfers remain partial when exported from saved history. Available receipt files are retained, the command exits nonzero and the error goes to stderr. Per-item failures also return a structured error with export metadata instead of emitting a success envelope. Safe `runs recover` reads can finish an interrupted export. A safe state read during recovery discards spent signed-upload capabilities after an observed start, including when the original start response was lost.

Compiled-process acceptance covers a server-side installation ACK whose response is lost: the candidate and proof survive, a new process verifies them, and the same installation converges without another registration. It also covers a lost bulk start response followed by a new process recovering the original task without another create, upload or start. SDK confirmation responses must have consistent generation metadata.

## Packages and documentation

Each repository has an explicit `release-files.json` allowlist. Package selection rejects traversal, duplicate paths and symlink components, and ignores stray local files. ZIP metadata is deterministic; Unix archive metadata is normalized. Dependency license/notice copies are shipped and linked from the READMEs. CLI build reports retain both runtime source fingerprints and the Go toolchain. The SDK package verifier cross-builds the packaged consumer for all five targets through an isolated local module proxy.

English is the default README with a top-level 简体中文 link. Chinese commands and executable SDK Go programs stay synchronized with English. Documentation distinguishes ID-only SDK examples from the CLI's private upload-capability journal, explains uncertain submissions, and keeps official FlashData links. Source and command reference generation are checked for drift.

## Before a public release

- Deploy the reviewed identity backend patch and verify the production onboarding feature configuration. Complete authorized live claim/handoff, callback and customer-storage checks.
- Run native Linux/macOS credential-store acceptance and the Linux race check. Windows native protection and subprocess workflows are covered locally; cross-compilation does not establish other native OS behavior.
- The CLI now pins the SDK commit through a Go pseudo-version in go.mod and go.sum, with no local replacement or sibling checkout. A future stable SDK tag can replace this pin after review; no version tag is needed for source builds or CI.
- Review the final source, archives and SHA-256 checksums; sign/notarize platform executables as appropriate to the distribution plan. No signing, remote CI execution or publication is claimed by this check.
