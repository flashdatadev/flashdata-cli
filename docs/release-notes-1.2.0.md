# FlashData CLI v1.2.0

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Set up your agent with two commands:

```sh
uv tool install flashdata-cli
flashdata setup
```

Choose Codex, Claude Code or Cursor and confirm the project directories. Setup installs 16 bundled CLI skills and verifies access without a paid query. No Node.js or separate skills installer is required. Agent skills use plain text by default; Markdown is optional and MCP remains optional.

- `setup` reuses existing access or initializes temporary access for a new profile. Default trial scopes are Google Search and YouTube metadata; choose other scopes before initialization or import an existing key with `auth login`.
- Use `--global` for a user-wide installation, or `--agent codex --non-interactive` for an explicit selection. `--skip-auth` installs skills offline.
- `skills list`, `skills install` and `skills verify` manage the bundled instructions without authentication. Edited and conflicting files are preserved and reported.
- Existing credentials, headless Linux login, run history and accepted-job recovery are preserved. The API, SDKs and ordinary CLI queries still default to JSON.

## Upgrade

```sh
uv tool upgrade flashdata-cli
flashdata setup
```

For pip, uv pip, Homebrew, signed APT/RPM or standalone archives, upgrade with your existing installer. See the [installation guide](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.md) and [setup options](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/agent-setup.md). Re-running setup refreshes unchanged installed skills and preserves local edits.

This release uses Go SDK v1.1.0 and includes five platform archives, documentation, a build report and SHA256SUMS. `flashdata --version` records the tagged source commit. Previous releases remain unchanged.

## 简体中文

通过 `uv tool install flashdata-cli` 安装，再运行 `flashdata setup`。选择 Codex、Claude Code 或 Cursor 并确认目录，即可安装随附的 16 个 CLI 技能并验证访问，不执行付费查询。无需 Node.js 或独立技能安装器。技能默认使用纯文本，Markdown 与 MCP 均为可选。

已有 uv 用户运行 `uv tool upgrade flashdata-cli` 后重新设置。`--global` 用于全局安装，`--skip-auth` 用于离线安装。已有凭据、运行历史及本地技能修改均会保留；默认试用范围为 Google 搜索和 YouTube 元数据。
