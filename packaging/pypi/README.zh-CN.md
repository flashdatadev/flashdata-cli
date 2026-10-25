# FlashData CLI

[English](https://github.com/flashdatadev/flashdata-cli/blob/main/packaging/pypi/README.md) | 简体中文

[FlashData](https://flashdata.dev) 为终端和 AI agent 提供 Google 搜索与 YouTube 数据。本包安装官方 [CLI](https://github.com/flashdatadev/flashdata-cli)，命令为 `flashdata`。Python SDK 是独立的 `flashdata` 软件包。

## 快速开始

```sh
uv tool install flashdata-cli
flashdata setup
```

选择 Codex、Claude Code 或 Cursor，并确认项目目录。设置会安装随附的 16 个 CLI 技能并验证访问，不执行付费查询。

**注册前即可试用。** 尚无凭据的新配置可申请临时访问，取决于试用可用性和资格，无需先注册账户或自行提供 API key。试用使用有限的共享 Credits，默认申请 Google 搜索和 YouTube 元数据。运行 `flashdata discover` 查看当前试用限制，`flashdata auth status` 检查访问。设置会复用已有凭据；已有密钥时可先运行 `flashdata auth login`，在隐藏提示中输入。

准备注册时，运行 `flashdata auth claim`，在浏览器中批准访问，再运行 `flashdata auth finish --wait`。重复设置不会替换已过期或耗尽额度的试用。

无需 Node.js 或独立技能安装器。技能默认使用纯文本，Markdown 与 MCP 可选。全局安装、选择技能及离线安装见[设置指南](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/agent-setup.zh-CN.md)；账户访问与其他试用范围见[接入指南](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/onboarding.md)。

试一次搜索：

```sh
flashdata search "battery recycling research" --num 5 --accept text/plain --json=false
```

查询消耗 Credits，通过 `flashdata account balance` 查看余额。只需终端命令时，改用 `flashdata init --name "My CLI"`，无需两个命令都运行。`setup --skip-auth` 仅离线安装技能，不修改凭据；查询仍需要已配置的访问。

普通 CLI 查询和 API 仍默认使用 JSON，详见[响应格式](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/response-formats.md)。

对于视频 ID 列表，`flashdata bulk run` 支持元数据、转录、字幕、许可标签和媒体下载。批量导出保持 JSONL，不添加 `--accept` 或 `--respond-with`。下载将媒体交付到你的云存储，并导出交付回执。配置、选项和恢复步骤见[批量工作流](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/bulk.md)。

## 其他安装方式

在已激活的 Python 3.10+ 虚拟环境中，选择 pip 或 uv pip：

```sh
python -m pip install --upgrade flashdata-cli
```

```sh
uv pip install --upgrade flashdata-cli
```

然后运行 `flashdata setup`。也可使用无需 Python 的原生 [Homebrew、APT、RPM 或独立下载包](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.zh-CN.md)。

## 升级

```sh
uv tool upgrade flashdata-cli
flashdata setup
```

pip 或 uv pip 使用 `--upgrade`。设置会更新未修改的技能，并保留本地修改。旧安装固定了精确版本时，可运行 `uv tool install --upgrade flashdata-cli` 替换版本约束。

若 uv 可执行目录不在 PATH 中，运行 `uv tool update-shell` 并重启终端。其他安装优先时，Bash 使用 `command -v flashdata`，PowerShell 使用 `Get-Command flashdata` 检查。

## 平台与包内容

版本 **1.2.1** 随附官方 CLI v1.2.1 可执行文件。启动器需要 Python 3.10+，没有 Python 运行时依赖，安装或启动时均不下载可执行文件。原生参数、输入输出及退出状态保持不变。

Wheel 支持 Linux glibc 2.17+ x86_64/ARM64、macOS 13+ Intel/Apple Silicon 以及 Windows x64。不提供 Linux musl/Alpine、Windows ARM64 或 32 位 wheel。离线安装可指定匹配平台的本地 wheel 文件。

按原安装方式使用 `uv tool uninstall flashdata-cli`、`python -m pip uninstall flashdata-cli` 或 `uv pip uninstall flashdata-cli` 卸载。卸载保留配置和执行历史；移除已保存凭据前请运行 `flashdata auth logout`。见[凭据存储](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/security.md)，包括无桌面 Linux 文件存储说明。

支持：[GitHub Issues](https://github.com/flashdatadev/flashdata-cli/issues) · [support@flashdata.dev](mailto:support@flashdata.dev)。
