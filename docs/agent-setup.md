# Guided agent setup

[FlashData](https://flashdata.dev) · [简体中文](agent-setup.zh-CN.md)

CLI 1.2.0 includes guided setup and the bundled skills. No Node.js or separate skills installer is needed. The PyPI update is pending; use a [native installer](installation.md#homebrew-linux-and-macos) in the meantime.

## Set up your agent

```sh
uv tool install flashdata-cli
flashdata setup
```

Select Codex, Claude Code or Cursor. The CLI suggests detected agents and asks you to confirm the target directories. It reuses existing access or initializes temporary access, installs the bundled CLI skills and verifies access and file contents. Verification makes no paid query. A new profile requests Google Search and YouTube metadata by default; use `--scopes` before initialization to choose other trial sources.

To choose the agent explicitly and install for your user account:

```sh
flashdata setup --agent codex --global --non-interactive
```

An existing key can be imported first with `flashdata auth login`. `--non-interactive` requires an explicit `--agent`; it never guesses where to install. `--skip-auth` installs and checks skills entirely offline, without creating or checking access.

## Install or verify skills separately

```sh
flashdata skills list
flashdata skills install --agent codex --skill google-search --skill youtube-transcript
flashdata skills verify --agent codex --skill google-search --skill youtube-transcript
```

The default selection contains 16 CLI skills. The MCP research skill is optional: select `--skill flashdata-mcp` to install its instructions. No command installs an MCP server or modifies your MCP configuration. Skills support plain text with `--accept text/plain --json=false`; Markdown requires `--respond-with markdown`.

| Agent | Project directory | User-wide directory (`--global`) |
| --- | --- | --- |
| Codex | `.agents/skills` | `~/.agents/skills` |
| Claude Code | `.claude/skills` | `~/.claude/skills`, or `$CLAUDE_CONFIG_DIR/skills` |
| Cursor | `.agents/skills` | `~/.cursor/skills` |

Use `--directory PATH` to select a project without changing directories. It cannot be combined with `--global`. Codex and Cursor share one project installation when selected together. Other agents can use the independent `npx skills` installer.

## Updates and recovery

Skills are bundled with the executable and pinned to the source revision reported by `skills list`. Upgrade the CLI, then rerun the same setup or install command. Installation uses no runtime download and requires no Node.js or Python.

The installer preserves local edits and conflicting same-name skills. It reports the path to review; back up or move the conflicting copy, then retry. Unrelated files remain untouched. Unchanged files previously installed by FlashData can be updated, and obsolete managed files removed. `skills verify` checks installed files against the current executable’s bundle.

If installation fails after access initialization, rerun the same command and profile; valid access is reused. An interrupted installation can be resumed. Remove `.flashdata-skills.lock` only after confirming no installer is running. Symlinked destinations are rejected; choose the real directory.

For credential storage, claiming and revoked access, see [onboarding](onboarding.md). Setup does not change query billing or retry paid work.
