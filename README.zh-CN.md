# FlashData CLI

[English](README.md) · [FlashData 官网](https://flashdata.dev) · [API 文档](https://flashdata.dev/docs) · [价格](https://flashdata.dev/pricing)

**在终端直接获取 Google 搜索结果和 YouTube 数据。** [FlashData](https://flashdata.dev) 通过统一 API 提供结构化搜索结果、视频元数据和已有字幕。使用 `flashdata` 研究主题、为 AI 助手提供网页信息、构建视频数据集，或将音频交付到自己的云存储。

无需编写应用代码。命令默认返回 JSON，并支持纯文本和 Markdown。批量任务导出 JSONL；保存的运行 ID 可以帮助脚本在中断后恢复已受理的任务。

## 按使用场景选择工作流

| 想要构建的功能 | 从这里开始 |
| --- | --- |
| 为 AI 助手提供网页上下文和来源链接 | [Google 搜索](#搜索网页用于研究或-ai-上下文) |
| 行业研究或新闻监控 | [新闻搜索](#搜索网页用于研究或-ai-上下文) |
| 按主题发现视频或学习资源 | [YouTube 搜索](#查找-youtube-视频并获取数据) |
| 视频分析和内容目录补全 | [视频元数据](#查找-youtube-视频并获取数据) |
| 为视频摘要、搜索或检索获取文本 | [已有字幕文本](#查找-youtube-视频并获取数据) |
| 根据一组视频 ID 构建数据集 | [批量元数据导出](#一条批量命令构建视频数据集) |
| 将音频或视频交付到自己的云存储桶 | [批量媒体下载](#完整批量工作流) |
| 继续中断的批量任务 | [恢复任务](#恢复中断的任务) |

以下示例负责获取数据；摘要、检索索引、调度和 AI 模型调用由应用实现。

## 快速开始

**注册前即可试用 FlashData。** 通过 [uv](https://docs.astral.sh/uv/) 安装，然后设置 agent：

```sh
uv tool install flashdata-cli
flashdata setup
```

引导设置需要 CLI 1.2.0 或更新版本。PyPI 更新发布前，请使用[原生安装器](docs/installation.zh-CN.md#homebrewlinux-和-macos)或[版本下载包](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.2.0)。

选择 Codex、Claude Code 或 Cursor，并确认项目目录。设置会安装随附的 16 个技能并验证访问，不执行付费查询。无需 Node.js 或独立技能安装器。Agent 技能默认请求纯文本结果，MCP 可选。

对于尚无凭据的新配置，设置会申请临时试用访问并保存在本地，无需先注册账户或自行提供 API key。试用取决于可用性和资格，提供有限的共享 Credits，默认可用 Google 搜索和 YouTube 元数据。运行 `flashdata discover` 查看当前试用限制，通过 `flashdata auth status` 检查访问状态。

设置会复用已有凭据。已有 API key 时，设置前运行 `flashdata auth login` 并在隐藏提示中输入。准备注册时，运行 `flashdata auth claim`，在浏览器中批准访问，再运行 `flashdata auth finish --wait`。详细步骤及其他试用来源见[接入指南](docs/onboarding.md)。

现在可以让 agent 搜索主题，或运行：

```sh
flashdata search "battery recycling startups" --num 5 --accept text/plain --json=false
```

查询消耗 Credits；设置本身不执行付费搜索。通过 `flashdata account balance` 查看剩余额度。重复设置不会替换已过期或耗尽额度的试用。只需终端命令时，改用 `flashdata init --name "My CLI"`，无需两个命令都运行。全局安装、选择技能，以及通过 `--skip-auth` 离线安装技能而不配置访问，见[设置指南](docs/agent-setup.zh-CN.md)。

### 升级与其他安装方式

已有 uv 用户运行 `uv tool upgrade flashdata-cli`，再运行 `flashdata setup` 更新随附技能。设置会保留本地修改并报告冲突。

也可使用 [pip 和 uv pip](docs/installation.zh-CN.md#pythonpip-和-uv)、[Homebrew](docs/installation.zh-CN.md#homebrewlinux-和-macos)、[APT](docs/installation.zh-CN.md#aptubuntu-和-debian)、[RPM](docs/installation.zh-CN.md#rpmfedora-和-rhel) 或[独立下载包](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.2.0)。原生包无需 Python、Go 或 Node.js。平台要求与 PATH 设置见[安装指南](docs/installation.zh-CN.md)。

## JSON、纯文本和 Markdown

上述所有安装方式均提供 1.2.0 正式版的 `--accept` 和 `--respond-with`。
也可使用 Go 1.25+ 从源码安装：

```bash
go install github.com/flashdatadev/flashdata-cli/cmd/flashdata@v1.2.0
```

为搜索选择纯文本或 Markdown：

```bash
flashdata search "vector databases" --accept text/plain
flashdata search "vector databases" --accept text/plain --respond-with markdown
```

JSON 仍是默认格式。`--accept` 也适用于后台回执、批量提交和任务状态。Markdown
适用于实时查询和已完成结果。增加 `--json` 可将文本放入 CLI 的 JSON 输出外层，
使用 `--output FILE` 可保存所选格式的正文。参见[响应格式与恢复](docs/response-formats.md)。

## 搜索网页，用于研究或 AI 上下文

将搜索结果保存为文件，供脚本或助手读取：

```bash
flashdata search "battery recycling startups" --num 5 --output search.json
```

通过 `--gl us` 和 `--hl en` 设置搜索国家及界面语言。新闻、图片、购物、视频或地点搜索可通过 `query` 指定对应来源。例如，将以下内容保存为 `news-query.json`：

```json
{"source":"google_news","query":"battery recycling"}
```

然后运行：

```bash
flashdata query --input news-query.json --output news.json
```

参阅[可用来源和输入参数](https://flashdata.dev/docs)，或运行 `flashdata sources list` 和 `flashdata sources show google_news`。

## 查找 YouTube 视频并获取数据

**按主题查找视频：**

```bash
flashdata youtube search "Go programming tutorial" --max-results 5 --json
```

**获取视频标题、频道、时长和互动数据：**

```bash
flashdata youtube metadata dQw4w9WgXcQ --output video.json
```

**获取已有的英文字幕，用于摘要或分析：**

```bash
flashdata youtube transcript dQw4w9WgXcQ --language en --output transcript.json
```

将 `dQw4w9WgXcQ` 替换为搜索结果中的视频 ID。字幕能否获取取决于该视频已有的语言轨道；此功能不会将任意音频转写为文字。如需已有的人工字幕轨道，使用 `youtube captions`。

**读取已有的人工字幕轨道：**

```bash
flashdata youtube captions dQw4w9WgXcQ --language en --output captions.json
```

`--json` 将机器可读响应写到标准输出；`--output FILE` 将结果保存为文件。每次导出请使用新的文件名：已有文件不会被覆盖。

## 一条批量命令构建视频数据集

创建 `ids.txt`，每行填写一个 YouTube 视频 ID：

```text
dQw4w9WgXcQ
jNQXAC9IVRw
```

获取所有 ID 的元数据并导出逐视频结果：

```bash
flashdata bulk run --source youtube_metadata --input ids.txt --run-id video-dataset-001 --timeout 1h --output videos.jsonl
```

此命令会创建任务、上传 ID、启动处理、等待完成，并写入 `videos.jsonl`。每行记录一个视频的处理状态和数据。将 `--source` 改为 `youtube_transcript`、`youtube_captions` 或 `youtube_trainability`，即可构建其他类型的批量数据集。选项和逐项状态说明见[批量处理指南](docs/bulk.md)。

随附的 `flashdata` 和五个 YouTube 批量技能提供 ID 文件准备、选项、任务控制和恢复指南。批量导出保持 JSONL 格式，包括下载交付回执；批量命令不添加 `--accept` 或 `--respond-with`。新的技能指南将在下一次 CLI 版本中发布。

## 完整批量工作流

### 将音频文件交付到云存储桶

使用相同的 `ids.txt` 为多个视频请求 MP3 音频。媒体文件会交付**到你的存储桶**；本地 JSONL 文件保存交付回执。

**1. 配置目标存储。** 在源码管理之外的私有工作目录中创建 `storage.json`，将所有占位符替换为自己的 S3 存储桶和凭据：

```json
{
  "storage_type": "s3",
  "storage_url": "YOUR_BUCKET_NAME",
  "region": "us-east-1",
  "access_key": "YOUR_STORAGE_ACCESS_KEY",
  "secret_key": "YOUR_STORAGE_SECRET_KEY",
  "object_prefix": "flashdata/audio"
}
```

在同一目录中创建 `audio-options.json`：

```json
{"download_type":"audio","output_format":"mp3"}
```

Linux/macOS 上，在运行命令前将两个文件设为私有：

```bash
chmod 600 storage.json audio-options.json
```

Windows 上，将两个文件的 ACL 限制为当前用户、SYSTEM 以及可选的 Administrators，详见[私有文件要求](docs/security.md)。还支持 S3 兼容存储、OSS、GCS、Azure、COS、BOS 和 TOS，参见[存储选项](docs/bulk.md)。

**2. 执行完整任务。** 在包含上述文件及 `ids.txt` 的目录中运行：

```bash
flashdata bulk run --source youtube_download --input ids.txt --storage-config storage.json --options audio-options.json --run-id audio-bulk-001 --timeout 1h --output audio-receipts.jsonl
```

命令会完成创建、上传、启动、等待和导出。查看 `audio-receipts.jsonl`，检查每个视频的交付状态和目标位置。即使任务已完成，也可能存在失败条目；回执会保留，命令会以非零状态退出。

**只处理一个视频**时，使用相同的存储配置：

```bash
flashdata youtube download dQw4w9WgXcQ --storage-config storage.json --format mp3 --wait --timeout 30m --output audio-delivery.json
```

### 将 MP4 视频交付到云存储桶

使用相同的 `ids.txt` 和私有 `storage.json`，创建 `video-options.json`：

```json
{"download_type":"audio_video","video_quality":"<=720","output_format":"mp4"}
```

Linux/macOS 上先限制文件权限：

```bash
chmod 600 video-options.json
```

```bash
flashdata bulk run --source youtube_download --input ids.txt --storage-config storage.json --options video-options.json --run-id video-bulk-001 --timeout 1h --output video-receipts.jsonl
```

视频文件写入存储桶，逐视频交付状态写入 `video-receipts.jsonl`。
Windows 上应使用与音频选项文件相同的私有 ACL。

### 恢复中断的任务

保持相同的配置和状态目录。列出保存的运行记录，然后将已受理的批量任务恢复到一个新的输出文件：

```bash
flashdata runs list
flashdata runs recover audio-bulk-001 --timeout 1h --output recovered-receipts.jsonl
```

恢复操作读取原有任务，不会提交另一次下载。恢复上面的元数据任务时，将运行 ID 换为 `video-dataset-001`；恢复 MP4 交付时使用 `video-bulk-001`。如果在上传或启动之前中断，请检查运行记录并遵循[批量恢复指南](docs/bulk.md)。超时只会停止本地等待，已受理的服务端任务仍会继续。

只有打算创建新任务时才使用新的运行 ID。脚本、结果不确定的提交和部分导出的说明见[运行历史与恢复](docs/runs.md)。

## 支持的 API

| 能力 | 命令 |
| --- | --- |
| Google 搜索、图片、新闻、购物、视频和地点 | `search`；全部六个来源均可使用 `query --input FILE` |
| 全部九个 YouTube 来源 | `youtube search/metadata/transcript/captions/download`；扩展搜索、频道、建议、许可标签和完整下载输入使用 `query` |
| 实时、后台和混合批次 | `query`、`query --mode async`、`batch submit` |
| 普通后台任务 | `jobs get`、`jobs wait`、`jobs result` |
| 全部五个批量来源、任务控制和结果 | `bulk run`、手动 `bulk` 命令、`runs recover` |
| 访问权限、余额、定价、限制和用量 | `auth status`、`account balance/pricing/limits/usage` |
| 公共来源目录和健康状态 | `sources list`、`discover` |
| 公共套餐目录和服务状态 | `discover plans`、`discover status` |

上述所有命令均包含在开箱即用的 **v1.2.0** 正式版中。`discover plans` 和 `discover status` 无需 API 密钥。已有安装可执行 `brew update && brew upgrade flashdatadev/tap/flashdata`、`sudo apt-get update && sudo apt-get install --only-upgrade flashdata` 或 `sudo dnf --refresh upgrade flashdata` 升级；软件源配置及 Windows 升级见[安装指南](docs/installation.zh-CN.md)。升级保留用户配置和执行历史。

全部 15 个来源支持后台/批次执行，其中 14 个支持实时请求。下载在后台运行，需要自己的云存储。
通过 `sources show SOURCE` 查看来源模式和可运行 JSON 请求。单视频 `youtube download` 快捷命令面向音频；`query --mode async` 和批量任务支持完整的媒体来源选项。

参阅[各 SDK 与 CLI 对比（英文）](docs/sdk-parity.md)和[命令参考](docs/commands.md)。

## 更多工作流

**消耗额度前先检查：** `--dry-run` 验证输入，不提交付费查询。余额查询同样免费。

```bash
flashdata search "battery recycling startups" --num 5 --dry-run
flashdata account balance
flashdata --help
```

- **脚本与 AI 代理：** 使用 `--json --non-interactive`、[配置和凭据管理](docs/security.md)，以及[稳定的运行 ID](docs/runs.md)。集成密钥管理器时可通过 `auth login --api-key-stdin` 保存密钥；如不希望保存密钥，可在每次执行命令时设置 `FLASHDATA_API_KEY`。环境变量中的密钥优先于已保存密钥；不会自动加载 `.env` 文件。
- **后台查询与混合批次：** 参阅[可运行输入示例](examples/README.md)及[命令参考](docs/commands.md)。
- **开始试用并连接账户：** 参阅[接入指南（英文）](docs/onboarding.md)，了解 `init`、浏览器授权及恢复操作。
- **在 Go 应用中使用 FlashData：** 安装本 CLI 所使用的 [Go SDK](https://github.com/flashdatadev/flashdata-go)。

[安装与升级](docs/installation.zh-CN.md) · [API 文档](https://flashdata.dev/docs) · [故障排查](docs/security.md) · [参与贡献与源码构建](CONTRIBUTING.md) · [版本说明](docs/release-notes-1.2.0.md) · [MIT 许可](LICENSE)
