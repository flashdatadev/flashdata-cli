# FlashData CLI v1.2.1

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

## Bulk workflows for agents

The bundled FlashData, YouTube metadata, transcript, captions, license-check and
download skills now include bulk workflows. They cover ID files, source options,
job controls, exports and recovery after an interruption. Each skill carries its
own guide and relevant example files.

Bulk exports remain JSONL, including per-item errors and download delivery
receipts. Ordinary skill content queries still use plain text by default; MCP is
optional. Bulk commands do not accept `--accept` or `--respond-with`.

Upgrade with your current installer, then run `flashdata setup` to refresh
unmodified skills. Setup preserves local edits and reports conflicts. Native
downloads, Homebrew, APT and RPM provide this release. PyPI wheels are prepared
for manual publication; until then, use a native installer.

## Clearer first-time setup

README, onboarding guides and CLI help now explain temporary access before
signup. For an eligible new profile, `flashdata setup` installs skills and requests
temporary access; terminal-only users can run `flashdata init --name "My CLI"`.
`--skip-auth` installs skills offline without changing credentials. Setup reuses
existing access and does not run a paid query.

See [installation](installation.md), [agent setup](agent-setup.md) and
[bulk workflows](bulk.md). Release checks exercise the bundled bulk examples
against mock APIs, including failed items and interrupted result transfers.

## 简体中文

随附的 FlashData、YouTube 元数据、转录、字幕、许可标签和下载技能新增批量工作流，
覆盖 ID 文件、来源选项、任务控制、导出及中断恢复。各技能自带指南与相关示例文件。
批量导出保持 JSONL 格式，保留逐项错误和下载交付回执。普通技能内容查询仍默认使用纯文本，
MCP 可选。批量命令不接受 `--accept` 或 `--respond-with`。

使用原安装器升级后，重新运行 `flashdata setup` 更新未修改的技能；本地修改会保留并提示冲突。
原生下载、Homebrew、APT 和 RPM 提供本版本。PyPI wheel 留待手动上传，上传前请使用原生安装器。

README、引导文档和命令帮助也说明了注册前使用临时访问的方式。符合条件的新配置可通过
`flashdata setup` 安装技能并请求临时访问；仅使用终端时运行 `flashdata init --name "My CLI"`。
`--skip-auth` 只离线安装技能，不更改凭据。设置会复用现有访问，不执行付费查询。
