# CLI setup

Install the standalone CLI with `uv tool install flashdata-cli`, or use the [installation guide](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.md) for native packages. Reuse an existing installation and access.

```sh
uv tool install flashdata-cli
flashdata setup
```

`setup` reuses valid access or saves temporary access for a new profile. Its default trial scopes are Google Search and YouTube metadata. For other sources, request the required scopes or connect an owner account; see [onboarding](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/onboarding.md). Existing API-key owners can use `flashdata auth login` with the hidden prompt. Never request a key in chat.

Check access without a paid query using `flashdata auth status`. On a new headless Linux profile, automatic credential storage can use an owner-only plaintext file when no desktop keyring is available. Existing encrypted profiles are never silently downgraded.

Setup requires CLI 1.2.0 or newer. Upgrade an existing uv installation with `uv tool upgrade flashdata-cli`, then rerun setup. Select Codex, Claude Code or Cursor and confirm the displayed directories. Use `flashdata skills install --agent AGENT` for installation without authentication. No Node.js or MCP package is needed.
