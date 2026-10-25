# FlashData CLI

English | [简体中文](https://github.com/flashdatadev/flashdata-cli/blob/main/packaging/pypi/README.zh-CN.md)

[FlashData](https://flashdata.dev) brings Google Search and YouTube data to your terminal and AI agent. This package installs the official [CLI](https://github.com/flashdatadev/flashdata-cli); the command is `flashdata`. The Python SDK is the separate `flashdata` package.

## Get started

```sh
uv tool install flashdata-cli
flashdata setup
```

Choose Codex, Claude Code or Cursor and confirm the project directories. Setup installs 16 bundled CLI skills and verifies access without a paid query.

**Try before signing up.** A new profile without credentials can request temporary access, subject to trial availability and eligibility. You do not need an account or an API key of your own. The trial uses limited shared Credits and requests Google Search and YouTube metadata by default. Run `flashdata discover` for current trial limits and `flashdata auth status` to check access. Setup reuses existing credentials; key owners can run `flashdata auth login` first and paste the key at its hidden prompt.

When ready to sign up, run `flashdata auth claim`, approve access in your browser, then run `flashdata auth finish --wait`. Rerunning setup does not replace an expired or exhausted trial.

No Node.js or separate skills installer is needed. Agent skills use plain text by default; Markdown and MCP are optional. See [setup options](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/agent-setup.md) for global installation, individual skills and offline setup, or [onboarding](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/onboarding.md) for account access and other trial scopes.

Try a search:

```sh
flashdata search "battery recycling research" --num 5 --accept text/plain --json=false
```

Queries use Credits; check what remains with `flashdata account balance`. For terminal-only use, run `flashdata init --name "My CLI"` instead of setup; you do not need both. `setup --skip-auth` only installs skills offline, leaving credentials unchanged; queries still need configured access.

Ordinary CLI queries and the API still default to JSON. See [response formats](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/response-formats.md).

For video-ID lists, `flashdata bulk run` supports metadata, transcripts, captions, license classifications and media downloads. Bulk exports stay JSONL; omit `--accept` and `--respond-with`. Downloads deliver media to your cloud storage and export delivery receipts. See [bulk workflows](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/bulk.md) for setup, options and recovery.

## Other installers

With Python 3.10+ in an active virtual environment, use pip or uv pip:

```sh
python -m pip install --upgrade flashdata-cli
```

```sh
uv pip install --upgrade flashdata-cli
```

Then run `flashdata setup`. Native [Homebrew, APT, RPM and standalone downloads](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.md) are also available and need no Python.

## Upgrade

```sh
uv tool upgrade flashdata-cli
flashdata setup
```

Use `--upgrade` for pip or uv pip. Setup updates unchanged skills and preserves local edits. If an older installation is pinned to an exact version, run `uv tool install --upgrade flashdata-cli` to replace that constraint.

If uv's executable directory is missing from PATH, run `uv tool update-shell` and restart the shell. Check `command -v flashdata` on Bash or `Get-Command flashdata` on PowerShell if another installation takes precedence.

## Platforms and package contents

Version **1.2.0** bundles the official CLI v1.2.0 executable. The launcher requires Python 3.10+, has no Python runtime dependencies and downloads no executable during installation or startup. Native arguments, streams and exit status pass through unchanged.

Wheels support Linux glibc 2.17+ x86_64/ARM64, macOS 13+ Intel/Apple Silicon, and Windows x64. No wheels are available for Linux musl/Alpine, Windows ARM64 or 32-bit systems. For offline installation, install the matching wheel from a local path.

Uninstall with `uv tool uninstall flashdata-cli`, `python -m pip uninstall flashdata-cli`, or `uv pip uninstall flashdata-cli`, matching your installer. Removing the package preserves profiles and run history; use `flashdata auth logout` first to remove a saved credential. See [credential storage](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/security.md), including headless Linux file storage.

Support: [GitHub Issues](https://github.com/flashdatadev/flashdata-cli/issues) · [support@flashdata.dev](mailto:support@flashdata.dev).
