# Changelog

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

## Unreleased

Add English/Chinese workflow selectors, uploaded-caption and MP4 bulk recipes, explicit API coverage and an SDK/CLI comparison. Add public discover plans/status commands on main, with no API key required. These two commands require a source build until the next binary release; published v1.0.0 assets are unchanged.

## Unreleased packaging

Local `flashdata-cli` 1.0.0 PyPI wheel candidates for Linux amd64/arm64, macOS 13+ Intel/Apple Silicon and Windows amd64, with a dependency-free Python launcher, pinned executable provenance, pip/uv installation checks and English/Chinese installation/build docs. No PyPI/TestPyPI publication or change to existing release binaries/tags.

## 1.0.0

Standalone CLI on the Go SDK; protocol-2 onboarding and fresh-process installation; native DPAPI/Keychain/Secret Service or explicit protected files; profiles and atomic/locked state; query/search/YouTube, async/batch and complete bulk workflows; durable run recovery; account/source discovery; diagnostics; legacy migration; English/简体中文 docs; offline tests and local archives/checksums. Formal GitHub Release with five platform archives, current English/Chinese installation docs, release notes, build report and checksums; no backend deployment.

Local review fixes preserve saved credentials under environment overrides and same-key imports; retain recurring policy during recovery; preserve valid public schemas; avoid per-item identity reads in batches; retain batch IDs after commit errors; protect Windows outputs from creation; enforce private credential-bearing inputs; recover large journals; refresh bulk stages and remove spent capabilities; and make migration retries safe. Added direct backend-contract checks, all-source example coverage, manual bulk JSON inputs and matching English/Chinese instructions.

Final joint check: retained candidate recovery IDs across SDK/CLI wire errors; partial-export errors; lost ACK/start recovery and spent upload cleanup; generation consistency; reviewed package manifests, deterministic archives, dependency notices and reproducible package checks. Source is hosted on GitHub; no version tag, GitHub Release or backend deployment is created by the source push.

The CLI pins SDK v1.0.0 and its verified checksums without a sibling replacement. Archives embed their exact source commit, include release notes and portable build reports, and ship SHA-256 checksums. CI runs native Intel macOS tests and cross-builds Apple Silicon on Go 1.25/1.27; superseded runs cancel automatically.
