# Agent 引导设置

[FlashData](https://flashdata.dev) · [English](agent-setup.md)

CLI 1.2.0 已包含引导设置与随附技能，无需 Node.js 或独立技能安装器。PyPI 更新发布前，请先使用[原生安装器](installation.zh-CN.md#homebrewlinux-和-macos)。

## 设置 agent

```sh
uv tool install flashdata-cli
flashdata setup
```

选择 Codex、Claude Code 或 Cursor。CLI 会提示检测到的 agent，并请你确认目标目录。它复用已有访问或初始化临时访问，安装随附 CLI 技能，并验证访问和文件内容。验证不执行付费查询。新配置默认请求 Google 搜索和 YouTube 元数据；首次初始化前可通过 `--scopes` 选择其他试用来源。

明确选择 agent 并安装到用户全局目录：

```sh
flashdata setup --agent codex --global --non-interactive
```

已有密钥可先通过 `flashdata auth login` 导入。`--non-interactive` 必须指定 `--agent`，不会猜测安装位置。`--skip-auth` 完全离线安装和检查技能，不创建或检查访问凭据。

## 单独安装或验证技能

```sh
flashdata skills list
flashdata skills install --agent codex --skill google-search --skill youtube-transcript
flashdata skills verify --agent codex --skill google-search --skill youtube-transcript
```

默认选择 16 个 CLI 技能。MCP 研究技能可通过 `--skill flashdata-mcp` 单独选择，只安装说明，不安装 MCP 服务或修改 MCP 配置。技能默认用 `--accept text/plain --json=false` 输出纯文本；Markdown 需要 `--respond-with markdown`。

| Agent | 项目目录 | 用户全局目录（`--global`） |
| --- | --- | --- |
| Codex | `.agents/skills` | `~/.agents/skills` |
| Claude Code | `.claude/skills` | `~/.claude/skills`，或 `$CLAUDE_CONFIG_DIR/skills` |
| Cursor | `.agents/skills` | `~/.cursor/skills` |

`--directory PATH` 指定项目目录，不能与 `--global` 同用。Codex 和 Cursor 在同一项目共用一份安装。其他 agent 可使用独立的 `npx skills` 安装器。

## 更新与恢复

技能随可执行文件分发，固定到 `skills list` 显示的源码版本。升级 CLI 后重新运行相同设置或安装命令。安装不下载远程文件，无需 Node.js 或 Python。

本地修改和同名冲突技能会保留并报告路径。先备份或移走冲突副本，再重试。其他文件不受影响；FlashData 上次安装且未被修改的文件可以更新，过期托管文件可以移除。`skills verify` 根据当前可执行文件随附内容验证安装。

若访问初始化后安装失败，使用相同命令和配置重试，已有有效访问会复用。安装中断后可以继续；只有确认没有安装器运行时才移除 `.flashdata-skills.lock`。不接受符号链接目标，请选择实际目录。

凭据、认领和撤销访问见[接入指南](onboarding.md)。设置不会改变查询计费或自动重试付费任务。
