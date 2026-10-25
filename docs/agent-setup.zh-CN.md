# Agent 引导设置

[FlashData](https://flashdata.dev) · [English](agent-setup.md)

CLI 1.2.0 已包含引导设置与随附技能，无需 Node.js 或独立技能安装器。PyPI 更新发布前，请先使用[原生安装器](installation.zh-CN.md#homebrewlinux-和-macos)。

## 设置 agent

```sh
uv tool install flashdata-cli
flashdata setup
```

选择 Codex、Claude Code 或 Cursor，并确认建议的目录。设置会安装随附 CLI 技能并验证访问，不执行付费查询。

**注册前即可试用。** 对于尚无凭据的新配置，设置会申请临时试用访问并保存在本地，无需自行提供 API key。试用取决于可用性和资格，已有凭据会被复用。默认试用来源为 Google 搜索和 YouTube 元数据；首次设置时可通过 `--scopes` 申请其他来源。安装技能不会增加数据来源的访问权限。

查询消耗有限的共享试用额度。运行 `flashdata discover` 查看当前试用限制，`flashdata auth status` 查看访问详情，`flashdata account balance` 查看剩余 Credits。准备注册时，运行 `flashdata auth claim`，在浏览器中批准访问，再运行 `flashdata auth finish --wait`。完整流程及恢复选项见[接入指南](onboarding.md)。重复设置不会替换已过期或耗尽额度的试用。

设置已包含 `init` 的访问配置步骤，无需两个命令都运行。只需终端命令、不安装 agent 技能时，改用 `flashdata init --name "My CLI"`。已有 API key 时，设置前通过 `flashdata auth login` 导入。

明确选择 agent 并安装到用户全局目录：

```sh
flashdata setup --agent codex --global --non-interactive
```

`--non-interactive` 必须指定 `--agent`，不会猜测安装位置。

### 仅安装技能，暂不配置访问

```sh
flashdata setup --agent codex --skip-auth
```

`--skip-auth` 完全离线安装和检查技能，不创建试用、不验证访问，也不修改凭据。查询仍需要已配置的访问。准备好后重新运行不带此参数的设置命令，或通过 `flashdata auth login` 导入已有密钥。

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

当前源码随附包中的主 FlashData 技能，以及元数据、转录、字幕、许可标签和下载技能均包含批量指南。批量导出保持 JSONL，普通技能内容查询默认使用纯文本。详见[批量工作流](bulk.md)。已发布的二进制文件将在下一次 CLI 发布时获得更新后的指南。

## 更新与恢复

技能随可执行文件分发，固定到 `skills list` 显示的源码版本。升级 CLI 后重新运行相同设置或安装命令。安装不下载远程文件，无需 Node.js 或 Python。

本地修改和同名冲突技能会保留并报告路径。先备份或移走冲突副本，再重试。其他文件不受影响；FlashData 上次安装且未被修改的文件可以更新，过期托管文件可以移除。`skills verify` 根据当前可执行文件随附内容验证安装。

若访问初始化后安装失败，使用相同命令和配置重试，已有有效访问会复用。安装中断后可以继续；只有确认没有安装器运行时才移除 `.flashdata-skills.lock`。不接受符号链接目标，请选择实际目录。

凭据、认领和撤销访问见[接入指南](onboarding.md)。设置不会改变查询计费或自动重试付费任务。
