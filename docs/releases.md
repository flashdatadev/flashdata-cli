# Releases, installation and source builds

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

The current release is [v1.2.0](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.2.0).
Install with `uv tool install flashdata-cli`, then run `flashdata setup` to choose
your agent, install bundled skills and verify access without a paid query. Agent
skills use plain-text output by default; MCP remains optional. See the [release notes](release-notes-1.2.0.md) and
[response formats](response-formats.md).

## Install

Standalone archives support Linux amd64/arm64, macOS Intel/Apple Silicon and
Windows x64, without a separate runtime. The [installation guide](installation.md)
and [简体中文](installation.zh-CN.md) cover direct downloads, Homebrew, signed APT
and signed RPM installation and upgrades. Preserve profiles and execution history
when updating. Existing release files and signed package versions remain immutable.

The CLI pins Go SDK **v1.1.0**, with its recorded Go checksums and no local module
replacement. Go 1.25+ users can also install
`github.com/flashdatadev/flashdata-cli/cmd/flashdata@v1.2.0` with `go install`.

## Build and verify

From a clean checkout of the intended tag, run
`python scripts/build_release.py --go /path/to/go --commit FULL_SOURCE_COMMIT`.
The script builds five CGo-disabled executables and archives with reviewed docs,
examples, licenses and a build report. It creates no tag and uploads no files.

The formal release contains the archives, both READMEs, `RELEASE_NOTES.md`,
`build-report.json` and `SHA256SUMS`. Executable `--version` output and the build
report record the exact tagged source commit. The report also records the pinned
SDK, source fingerprints and toolchain. Verify the canonical tag and every asset
checksum before publishing; keep existing tags and released bytes unchanged.

[GitHub Actions](https://github.com/flashdatadev/flashdata-cli/actions) runs Linux,
Windows and Intel macOS source checks on Go 1.25/1.27, Linux race checks, source
and command generation checks, archive-policy tests and all five native archive
acceptance jobs. Release checks use mock APIs and submit no paid queries.

The v1.2.0 wheels for [PyPI](https://pypi.org/project/flashdata-cli/) are prepared for
manual publication. Until they are uploaded, use a native installer for guided setup.
The CLI installer is separate from the Python SDK. See
[Python installation](installation.md#python-pip-and-uv) and
[wheel preparation](python-packaging.md).
