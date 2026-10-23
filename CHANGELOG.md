# Changelog

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

## Unreleased

Make default login and initialization work on fresh headless Linux profiles even
when a session bus is configured but the desktop keyring is unavailable. Use
owner-only file storage in that case, and reuse it on later commands. Existing
keyring profiles and explicit `--credential-store os` still require the keyring.

## 1.1.1 — 2026-10-23

Fix consecutive commands with `FLASHDATA_API_KEY` on a new profile. Save profile
origins before the first run without saving the environment key, so later queries,
account commands and recovery work with the same environment variable. Preserve
saved credentials, origin bindings and protection against duplicate submissions.

## 1.1.0 — 2026-10-23

Redact credentials escaped by the API's Markdown renderer in JSON and plain-text
responses, including saved results and file exports. Preserve the last saved
response when a format change cannot be committed, and retain safe recovery after
an interruption. Cover both async and batch recovery with write-failure regressions.

Add independent `--accept` and `--respond-with` flags for API responses, preserving
JSON defaults and the CLI's `--json` envelope. Support plain-text receipts with
durable job IDs, Markdown completed results, cached-format checks and safe
background-result reformatting. Pin Go SDK v1.1.0, update help and
English/Chinese documentation, and test all sources against API-generated fixtures.

## 1.0.1

Ship public `discover plans` and `discover status` commands in all five standalone binaries; neither needs an API key. Add English/Chinese workflow selectors, uploaded-caption and complete MP3/MP4 bulk recipes, explicit API coverage and an SDK/CLI comparison. Update release installation and upgrade instructions for Homebrew, APT and RPM. Native archive acceptance tests run the packaged binaries against mock APIs on every supported platform. The Go SDK stays pinned to v1.0.0; existing v1.0.0 release assets are unchanged.

## Unreleased packaging

Local `flashdata-cli` 1.0.0 PyPI wheel candidates for Linux amd64/arm64, macOS 13+ Intel/Apple Silicon and Windows amd64, with a dependency-free Python launcher, pinned executable provenance, pip/uv installation checks and English/Chinese installation/build docs. No PyPI/TestPyPI publication or change to existing release binaries/tags.

## 1.0.0

Standalone CLI on the Go SDK; protocol-2 onboarding and fresh-process installation; native DPAPI/Keychain/Secret Service or explicit protected files; profiles and atomic/locked state; query/search/YouTube, async/batch and complete bulk workflows; durable run recovery; account/source discovery; diagnostics; legacy migration; English/简体中文 docs; offline tests and local archives/checksums. Formal GitHub Release with five platform archives, current English/Chinese installation docs, release notes, build report and checksums; no backend deployment.

Local review fixes preserve saved credentials under environment overrides and same-key imports; retain recurring policy during recovery; preserve valid public schemas; avoid per-item identity reads in batches; retain batch IDs after commit errors; protect Windows outputs from creation; enforce private credential-bearing inputs; recover large journals; refresh bulk stages and remove spent capabilities; and make migration retries safe. Added direct backend-contract checks, all-source example coverage, manual bulk JSON inputs and matching English/Chinese instructions.

Final joint check: retained candidate recovery IDs across SDK/CLI wire errors; partial-export errors; lost ACK/start recovery and spent upload cleanup; generation consistency; reviewed package manifests, deterministic archives, dependency notices and reproducible package checks. Source is hosted on GitHub; no version tag, GitHub Release or backend deployment is created by the source push.

The CLI pins SDK v1.0.0 and its verified checksums without a sibling replacement. Archives embed their exact source commit, include release notes and portable build reports, and ship SHA-256 checksums. CI runs native Intel macOS tests and cross-builds Apple Silicon on Go 1.25/1.27; superseded runs cancel automatically.
