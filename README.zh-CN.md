# FlashData CLI

[English](README.md) · [API 文档](https://flashdata.dev/docs) · [价格](https://flashdata.dev/pricing)

在终端和 AI agent 中使用 Google 搜索与 YouTube 数据。搜索网页、获取视频元数据和转录文本，或通过批量任务处理视频列表。

## 快速开始

使用 [uv](https://docs.astral.sh/uv/) 安装，然后设置 agent：

```sh
uv tool install flashdata-cli
flashdata setup
```

选择 Codex、Claude Code 或 Cursor，并确认项目目录。设置会安装技能并配置访问，不执行付费查询。

**注册前即可试用。** 符合条件的新用户可获得临时访问及有限 Credits，默认用于 Google 搜索和 YouTube 元数据。设置会复用已有凭据。已有密钥时，设置前运行 `flashdata auth login`。试用限制和后续注册步骤见[账户接入](docs/onboarding.md)。

运行第一次搜索：

```bash
flashdata search "battery recycling startups" --num 5
```

只使用终端时，运行 `flashdata init --name "My CLI"` 代替 `setup`。数据查询消耗 Credits；通过 `flashdata account balance` 查看剩余额度。

<details>
<summary>其他安装方式</summary>

在已激活的 Python 3.10+ 虚拟环境中，选择 pip 或 uv pip：

```sh
python -m pip install --upgrade flashdata-cli
```

```sh
uv pip install --upgrade flashdata-cli
```

macOS/Linux 也可使用 Homebrew：

```sh
brew install flashdatadev/tap/flashdata
```

安装后运行 `flashdata setup`。也支持 [APT、RPM 和独立下载包](docs/installation.zh-CN.md)。原生包无需 Python、Go 或 Node.js。

</details>

## 常用命令

保存 Google 搜索结果：

```bash
flashdata search "battery recycling startups" --num 5 --output search.json
```

搜索 YouTube 视频，再使用返回的视频 ID 获取元数据或已有转录文本：

```bash
flashdata youtube search "Go programming tutorial" --max-results 5
flashdata youtube metadata dQw4w9WgXcQ --output video.json
flashdata youtube transcript dQw4w9WgXcQ --language en --output transcript.json
```

请替换示例视频 ID。转录文本是否可用取决于视频的语言轨道；上传者字幕使用 `youtube captions` 获取。

需要 Google 新闻、图片、购物、视频、地点或其他 YouTube 来源时，运行 `flashdata sources list`，再通过 `flashdata sources show SOURCE` 查看输入格式。使用 `flashdata query --input query.json` 提交请求。详见[命令参考](docs/commands.md)和[请求示例](examples/README.md)。

## 输出格式

CLI 查询默认返回 JSON。需要时可选择纯文本或 Markdown：

```bash
flashdata search "vector databases" --accept text/plain
flashdata search "vector databases" --accept text/plain --respond-with markdown
```

`--output FILE` 将所选响应正文保存到新文件；`--json` 将文本放入 CLI 的 JSON 响应结构。[Agent 技能](https://github.com/flashdatadev/flashdata-skills)默认请求纯文本；批量导出始终使用 JSONL。异步任务和批次的格式选项见[响应格式](docs/response-formats.md)。

## 批量任务

创建 `ids.txt`，每行一个 YouTube 视频 ID：

```text
dQw4w9WgXcQ
jNQXAC9IVRw
```

导出列表中视频的元数据：

```bash
flashdata bulk run --source youtube_metadata --input ids.txt --run-id video-dataset-001 --timeout 1h --output videos.jsonl
```

该命令提交任务、等待完成，并为每个视频保存一行 JSONL。批量任务也支持转录文本、字幕、许可分类和下载。下载的媒体文件进入你的云存储，JSONL 导出保存交付回执。

中断后，将同一任务的结果恢复到新文件：

```bash
flashdata runs recover video-dataset-001 --timeout 1h --output recovered-videos.jsonl
```

恢复会读取已有任务，不会重新提交。MP3/MP4 示例、存储配置、失败项及任务控制见[批量指南](docs/bulk.md)。

## 升级

```sh
uv tool upgrade flashdata-cli
flashdata setup
```

设置会更新未修改的技能，并保留本地修改。其他安装方式见[升级指南](docs/installation.zh-CN.md)。

## 更多

- [Agent 设置](docs/agent-setup.zh-CN.md)：选择技能、全局安装，或用 `--skip-auth` 离线安装技能而不改变凭据。MCP 可选。
- [账户接入](docs/onboarding.md)：临时访问、API key 和账户连接。
- [请求示例](examples/README.md)：后台查询、批次及完整来源选项。
- [凭据与自动化](docs/security.md) · [任务恢复](docs/runs.md) · [参与贡献](CONTRIBUTING.md)

[反馈问题](https://github.com/flashdatadev/flashdata-cli/issues) · [版本发布](https://github.com/flashdatadev/flashdata-cli/releases) · [MIT 许可证](LICENSE)
