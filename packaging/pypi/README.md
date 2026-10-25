# FlashData CLI

[简体中文](https://github.com/flashdatadev/flashdata-cli/blob/main/packaging/pypi/README.zh-CN.md) · [GitHub](https://github.com/flashdatadev/flashdata-cli) · [API docs](https://flashdata.dev/docs)

Google Search and YouTube data for your terminal and AI agent. This package installs the `flashdata` command. Requires Python 3.10+; the Python SDK is the separate [`flashdata` package](https://pypi.org/project/flashdata/).

## Get started

```sh
uv tool install flashdata-cli
flashdata setup
```

Choose Codex, Claude Code or Cursor. Setup installs the skills and configures access without a paid query. Eligible new users can try Google Search and YouTube metadata with temporary access and limited Credits before signing up. Existing credentials are reused; if you have an API key, run `flashdata auth login` first.

```sh
flashdata search "battery recycling research" --num 5
```

For terminal-only use, run `flashdata init --name "My CLI"` instead of setup. See [account setup](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/onboarding.md) for trial limits and signing up later.

## Install with pip or uv pip

In an active virtual environment, choose one:

```sh
python -m pip install --upgrade flashdata-cli
```

```sh
uv pip install --upgrade flashdata-cli
```

Then run `flashdata setup`. Native [Homebrew, APT, RPM and standalone downloads](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.md) are also available.

## Output and bulk jobs

CLI queries return JSON by default. Use `--accept text/plain` for text and add `--respond-with markdown` for Markdown. Agent skills request plain text by default; MCP is optional.

Bulk jobs process video-ID lists for metadata, transcripts, captions, license classifications and downloads. Exports stay JSONL. Downloads deliver media to your cloud storage and export delivery receipts. See [commands and examples](https://github.com/flashdatadev/flashdata-cli#common-commands) and the [bulk guide](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/bulk.md).

## Upgrade

```sh
uv tool upgrade flashdata-cli
flashdata setup
```

For pip or uv pip, repeat the install command with `--upgrade`. Setup refreshes unmodified skills and preserves local edits.

## Supported platforms

- Linux glibc 2.17+: x86_64 and ARM64.
- macOS 13+: Intel and Apple Silicon.
- Windows: x64.

For PATH setup, offline installation and removal, see the [installation guide](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.md).

[Report an issue](https://github.com/flashdatadev/flashdata-cli/issues) · [Pricing](https://flashdata.dev/pricing) · [Agent setup](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/agent-setup.md)
