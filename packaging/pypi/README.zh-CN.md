# FlashData CLI

[English](https://github.com/flashdatadev/flashdata-cli/blob/main/packaging/pypi/README.md) · [GitHub](https://github.com/flashdatadev/flashdata-cli) · [API 文档](https://flashdata.dev/docs)

在终端和 AI agent 中使用 Google 搜索与 YouTube 数据。本包安装 `flashdata` 命令，需要 Python 3.10+；Python SDK 是独立的 [`flashdata` 包](https://pypi.org/project/flashdata/)。

## 快速开始

```sh
uv tool install flashdata-cli
flashdata setup
```

选择 Codex、Claude Code 或 Cursor。设置会安装技能并配置访问，不执行付费查询。符合条件的新用户可在注册前通过临时访问和有限 Credits 试用 Google 搜索及 YouTube 元数据。已有凭据会被复用；已有 API key 时，先运行 `flashdata auth login`。

```sh
flashdata search "battery recycling research" --num 5
```

只使用终端时，运行 `flashdata init --name "My CLI"` 代替设置。试用限制和后续注册步骤见[账户接入](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/onboarding.md)。

## 使用 pip 或 uv pip 安装

在已激活的虚拟环境中，选择一种方式：

```sh
python -m pip install --upgrade flashdata-cli
```

```sh
uv pip install --upgrade flashdata-cli
```

安装后运行 `flashdata setup`。也可使用原生 [Homebrew、APT、RPM 和独立下载包](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.zh-CN.md)。

## 输出与批量任务

CLI 查询默认返回 JSON。使用 `--accept text/plain` 选择纯文本，再添加 `--respond-with markdown` 选择 Markdown。Agent 技能默认请求纯文本，MCP 可选。

批量任务支持视频列表的元数据、转录文本、字幕、许可分类和下载，导出保持 JSONL。下载的媒体文件进入你的云存储，导出文件保存交付回执。详见[命令与示例](https://github.com/flashdatadev/flashdata-cli/blob/main/README.zh-CN.md#常用命令)和[批量指南](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/bulk.md)。

## 升级

```sh
uv tool upgrade flashdata-cli
flashdata setup
```

pip 或 uv pip 用户可重复执行带 `--upgrade` 的安装命令。设置会更新未修改的技能，并保留本地修改。

## 支持的平台

- Linux glibc 2.17+：x86_64 和 ARM64。
- macOS 13+：Intel 和 Apple Silicon。
- Windows：x64。

PATH 设置、离线安装和卸载方式见[安装指南](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.zh-CN.md)。

[反馈问题](https://github.com/flashdatadev/flashdata-cli/issues) · [价格](https://flashdata.dev/pricing) · [Agent 设置](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/agent-setup.zh-CN.md)
