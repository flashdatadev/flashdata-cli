# FlashData CLI

[English](README.md) · [FlashData 官网](https://flashdata.dev) · [API 文档](https://flashdata.dev/docs) · [价格](https://flashdata.dev/pricing) · [v1.0.0 正式版](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.0)

面向开发者、编程代理和个人代理的独立 `flashdata` 可执行程序。[FlashData](https://flashdata.dev) 提供结构化 Google 搜索、YouTube 数据、已有转录/字幕、后台任务及客户存储媒体交付。CLI 基于 Go SDK，负责安全配置、可恢复接入、持久化执行记录和流式 Bulk 导出。

**版本 1.0.0。** 可从 [v1.0.0 正式版](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.0)下载即用型可执行程序。 支持 Linux amd64/arm64、macOS Intel/Apple Silicon、Windows amd64。安装二进制后不需要 Node.js、Python、npm、Docker 或单独的 Go 运行时；源码开发需要 Go 1.25+。

## 安装 v1.0.0

### Python 安装器（本地候选包）

**尚未发布到 PyPI。** 本地平台 wheel 支持 `pip`、`uv pip` 和 `uv tool`，使用同一份已核验的 CLI 二进制。这种安装方式需要 Python 3.10+，无需安装 Go。构建说明及平台范围见[英文指南](docs/python-packaging.md)和[简体中文](docs/python-packaging.zh-CN.md)。

构建 `dist/wheels` 后，从仓库根目录任选一种方式。pip/uv pip 需要已激活的虚拟环境：

```bash
python -m pip install --no-index --find-links dist/wheels flashdata-cli==1.0.0
# Or, inside a virtual environment:
uv pip install --no-index --find-links dist/wheels flashdata-cli==1.0.0
# Or, as an isolated CLI tool:
uv tool install --no-index --find-links dist/wheels flashdata-cli==1.0.0
```

安装后使用 `flashdata` 命令。[安装、重装与卸载](docs/installation.zh-CN.md#python-安装器本地候选包)介绍了具体操作。发行包 `flashdata-cli` 与 Python SDK `flashdata` 相互独立；经过明确授权的 PyPI 正式发布后才能直接从注册表安装。

### Homebrew（Linux 和 macOS）

安装 [Homebrew](https://brew.sh) 并配置 PATH 后执行：

```bash
brew install flashdatadev/tap/flashdata
```

[FlashData 官方软件源](https://github.com/flashdatadev/homebrew-tap)会自动选择 Linux amd64/arm64 或 macOS Intel/Apple Silicon 版本，并验证归档的 SHA-256 校验和，无需安装 Go。升级、卸载、附带示例及从手动安装迁移的方法见[安装指南](docs/installation.zh-CN.md#homebrewlinux-和-macos)。

### APT（Ubuntu 和 Debian）

先完成[签名软件源的一次性配置](docs/installation.zh-CN.md#aptubuntu-和-debian)，然后安装：

```bash
sudo apt-get update
sudo apt-get install flashdata
```

[官方 APT 软件源](https://github.com/flashdatadev/apt)支持 Ubuntu 22.04/24.04 与 Debian 12/13，架构为 amd64/arm64。软件包 **1.0.0-1** 包含 CLI **1.0.0**。配置过程会验证专用签名密钥指纹，并通过 `Signed-By` 限定信任范围。升级、卸载及从其他安装方式迁移的方法见[安装指南](docs/installation.zh-CN.md#aptubuntu-和-debian)。

### RPM（Fedora 和 RHEL）

先完成[签名软件源的一次性配置](docs/installation.zh-CN.md#rpmfedora-和-rhel)，然后安装：

```bash
sudo dnf --refresh install flashdata
```

[官方 RPM 软件源](https://github.com/flashdatadev/rpm)提供 x86_64/aarch64 的 **1.0.0-1** 软件包。原生 CI 覆盖 Fedora 43/44，并通过 Red Hat UBI 容器覆盖 RHEL 8/9/10 用户空间。配置过程核对独立 RPM 签名证书的完整指纹，软件包和仓库元数据均有签名。升级、信任机制及从其他安装方式迁移的方法见[安装指南](docs/installation.zh-CN.md#rpmfedora-和-rhel)。

### 独立安装包

下载与操作系统和 CPU 匹配的安装包，无需安装 Go。

| 平台 | 下载 |
| --- | --- |
| Linux amd64 | [flashdata_1.0.0_linux_amd64.tar.gz](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_linux_amd64.tar.gz) |
| Linux arm64 | [flashdata_1.0.0_linux_arm64.tar.gz](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_linux_arm64.tar.gz) |
| macOS Intel | [flashdata_1.0.0_darwin_amd64.tar.gz](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_darwin_amd64.tar.gz) |
| macOS Apple Silicon | [flashdata_1.0.0_darwin_arm64.tar.gz](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_darwin_arm64.tar.gz) |
| Windows amd64 | [flashdata_1.0.0_windows_amd64.zip](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_windows_amd64.zip) |

先对照 [SHA256SUMS](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/SHA256SUMS) 校验下载文件，再解压并将可执行程序放入 PATH 目录。[安装指南](docs/installation.zh-CN.md)提供可直接运行的 Linux/macOS 与 Windows 命令，包含校验和验证及 PATH 设置。

```bash
flashdata --version
flashdata --help
```

正式版本还提供[版本说明](docs/release-notes-1.0.0.md)、构建记录及中英文 README。

## 从源码构建

克隆此版本并使用 Go 1.25+ 构建。依赖锁定为 SDK **v1.0.0**，并由 `go.sum` 校验；无需相邻的 SDK 仓库或本地替换。详见 [v1.0.0 版本说明](docs/release-notes-1.0.0.md)。

```bash
git clone --branch v1.0.0 --depth 1 https://github.com/flashdatadev/flashdata-cli.git
cd flashdata-cli
go mod download
go build -trimpath -o dist/flashdata ./cmd/flashdata
./dist/flashdata --version
./dist/flashdata --help
```

PowerShell：

```powershell
go build -trimpath -o dist/flashdata.exe ./cmd/flashdata
.\dist\flashdata.exe --version
.\dist\flashdata.exe --help
```

将匹配平台的可执行文件复制到用户自有的 PATH 目录，命名为 `flashdata`，Windows 使用 `flashdata.exe`。本地打包脚本生成归档和 SHA-256 校验文件，不上传任何内容。GitHub Release 的正式安装包可公开下载，无需登录，详见[发布准备](docs/releases.md)。

## 已有账户：导入 API 密钥

从 [FlashData 账户](https://flashdata.dev) 获取密钥。终端交互输入会隐藏字符：

```bash
flashdata auth login
```

代理、CI 或其他无终端环境必须通过标准输入提供一条密钥：

```text
your-secret-manager read flashdata-api-key | flashdata auth login --api-key-stdin --non-interactive
```

上例的密钥管理器命令仅为示意，请替换为真实工具。密钥不能放在命令参数或调度器文本中。显式设置的 `FLASHDATA_API_KEY` 仅覆盖当前命令的已保存密钥，不会被隐式写入凭据库；`auth login` 才是明确导入操作。不会自动加载 `.env`。

```bash
flashdata auth status
flashdata sources list
flashdata search "latest battery recycling research" --num 5 --run-id research-001
```

提交查询、后台任务时遵循 FlashData 正常计费规则；公开发现、账户读取及 doctor 不提交付费查询。

## 新代理：接入、浏览器认领与安装确认

```bash
flashdata discover
flashdata init --name "Research assistant" --client codex --scopes google_search,youtube_metadata
flashdata auth claim
# Open the returned URL, enter the separately displayed code and approve in your browser.
flashdata auth finish --wait --timeout 10m
flashdata auth status
```

`init` 验证并复用已有访问。仅在配置从未建立、且提供明确 `--name` 时注册临时连接。注册证明和精确输入在 POST 前持久化；响应丢失后使用相同证明恢复。过期、撤销、无效密钥、缺失凭据文件、余额耗尽或已退出的配置都不会触发替代试用。

初始试用额度、按价格计算的请求当量与当前余额分别展示。不同数据源的当量是同一共享初始额度的替代用法，不是独立配额。实际可用性、范围、额度与到期时间以服务端为准。

浏览器批准后，`auth finish` 原子保存候选密钥，以全新进程从持久化存储读取并验证，然后才发送 ACK。验证失败时保留候选与证明，不发送 ACK；服务端确认 installed 后清理过时的注册、setup 和 handoff 证明。直接 REST 使用普通 API 密钥，不使用 MCP OAuth 令牌。详见[接入指南](docs/onboarding.md)。

接入能力受服务端开关控制。导入密钥的周期性执行需要配套本地后端补丁提供的身份端点；身份不可用时不会猜测所有权。已认领连接可使用现有的服务端连接身份验证。此次未进行任何后端部署。

重复导入同一已保存密钥会保留接入证明和上传恢复信息；`--replace` 才明确丢弃这些信息。Bulk 命令使用环境密钥覆盖时会保留原有已保存凭据；向仅使用环境密钥的配置显式导入密钥也会保留已接收任务的恢复信息。

## JSON 查询、预演和输出

```bash
flashdata query --input examples/query.json --dry-run
flashdata query --input examples/query.json --run-id research-002 --output research.json
flashdata youtube metadata dQw4w9WgXcQ
flashdata youtube transcript dQw4w9WgXcQ --language en
flashdata youtube captions dQw4w9WgXcQ --language en
```

支持文件输入及 `--input -` 标准输入，规范格式如下：

```json
{"source":"google_search","query":"latest battery recycling research","params":{"num":5}}
```

也支持 CLI 规范中的平铺形式，即顶层 `"num":5`，在哈希和提交前规范化为 `params.num`。未知字段、重复 JSON 字段、平铺/嵌套冲突及无效参数会被拒绝，不会擅自补充查询默认值。含云凭据或回调 URL 的 query/batch 文件必须具有私有权限，也可通过标准输入读取。转录与字幕快捷命令分别显式设置 `auto:true`、`auto:false`，语言默认为 en，可通过参数修改。

`--dry-run` 进行本地 SDK 校验、实时访问验证及可用价格估算，不提交付费查询或创建 Bulk 任务；估算不是服务端预算限制。结果保留来源 URL、时间戳、额外字段、计费信息及大整数；检索文本仅作为数据处理。

无终端时默认输出版本化 JSON，`--json` 强制机器输出；终端使用缩进 JSON。成功写 stdout，错误和诊断写 stderr。成功退出码 0，失败 1。`--output` 不覆盖现有文件；异步未等待时保存准入回执，指定 `--wait` 才等待结果。大批量 JSONL 留在私有记录目录中，通过文件导出，避免构造巨大的 JSON 信封。

```json
{"schemaVersion":1,"ok":true,"runId":"research-001","cached":false,"data":{},"usage":{"credits":0.02}}
```

示例额度仅用于说明，实际 usage 来自 API。

## 周期性任务与恢复

调度由宿主负责，每个计划执行实例必须提供稳定且唯一的 run ID：

```bash
flashdata query --input examples/query.json --recurring --run-id daily-research:2026-10-22:Asia-Shanghai --profile research --json --non-interactive
```

请先导入或认领对应 profile。周期性任务需要当前服务端确认的所有权；认领密钥还需要安装 ACK。相同配置、相同 ID、相同规范化输入，在重新验证权限后返回缓存结果，重启进程后也成立。输入不同则冲突，账户或服务来源不同不能读取旧数据。使用 `--recurring` 创建的记录在后续缓存读取和恢复时仍要求已验证所有权与安装状态，即使省略该参数也不会绕过。

```bash
flashdata runs list
flashdata runs show research-001
flashdata runs show research-001 --result --output recovered.json
flashdata query --input examples/youtube-metadata.json --mode async --run-id metadata-001
flashdata runs recover metadata-001 --output metadata.json
flashdata jobs get job_YOUR_ID
flashdata jobs wait job_YOUR_ID --timeout 10m
flashdata jobs result job_YOUR_ID --output job-result.json
flashdata batch submit --input examples/batch.json --run-id batch-001 --wait
```

执行意图在提交前保存。丢失响应属于 uncertain，失败/不确定任务绝不自动重投。`runs recover` 仅读取已知远程任务，不重复创建、查询、上传缺失输入或恢复暂停。超时和取消仅停止本地等待，服务端任务可能继续。一键命令可自动生成 ID；不复用 ID 就代表新的执行。本地去重仅在同一持久化目录有效，不承诺跨机器 exactly-once。详见[执行记录](docs/runs.md)。

响应格式异常时仍能恢复出的有效任务 ID，会以 `candidateRemoteIds` 和 `acceptanceConfirmed:false` 保留在错误与持久化记录中。使用 `jobs get` 或 `bulk get` 核对这些候选 ID，不能把不完整的批次当成整体提交成功。

## 完整 Bulk 流程

元数据、转录、字幕、许可标签和下载使用同一持久化流程。先免费预演，再创建、上传 ID、启动、等待并导出回执：

```bash
flashdata bulk run --source youtube_metadata --input examples/ids.txt --run-id metadata-bulk-001 --dry-run
flashdata bulk run --source youtube_metadata --input examples/ids.txt --run-id metadata-bulk-001 --timeout 1h --output metadata-receipts.jsonl
```

媒体下载需要在仓库外的私有文件配置真实客户存储，可参考 [storage.example.json](examples/storage.example.json)。音频选项可放入私有文件 `{"download_type":"audio","output_format":"mp3"}`。云凭据不能出现在参数、调度器文本或已跟踪文件中。Unix 文件必须由当前用户拥有且权限为 0600；Windows ACL 只允许当前用户、SYSTEM 及可选 Administrators。

```bash
flashdata bulk run --source youtube_download --input examples/ids.txt --storage-config /private/flashdata/storage.json --options /private/flashdata/download-options.json --run-id audio-bulk-001 --timeout 1h --output audio-receipts.jsonl
# Recover accepted work without another creation/upload:
flashdata runs recover audio-bulk-001 --timeout 1h --output recovered-receipts.jsonl
```

媒体二进制直接写入客户存储桶，JSONL 保存每项结果/交付回执。整个任务 completed 仍可能包含 failed/invalid 条目；CLI 先保留回执，再以非零状态退出。部分流也保留，导出文件不会被覆盖。支持 S3、S3 兼容、OSS、GCS、Azure、COS、BOS、TOS，真实权限及提供方配置由服务端验证。

手动创建示例使用仓库内不含秘密的元数据配置。以下为 Bash 命令；将 `bulk_YOUR_ID` 替换为 create 返回的 ID，并在上传、启动前修改选项：

```bash
flashdata bulk create --input - --run-id bulk-manual-001 < examples/bulk-create.json
flashdata bulk update bulk_YOUR_ID --input - < examples/bulk-update.json
flashdata bulk upload bulk_YOUR_ID --input examples/ids.txt --run-id bulk-manual-001
flashdata bulk start bulk_YOUR_ID
flashdata bulk get bulk_YOUR_ID
flashdata bulk wait bulk_YOUR_ID --timeout 1h
flashdata bulk results bulk_YOUR_ID --run-id bulk-read-001 --output receipts.jsonl
```

PowerShell 的标准输入写法如下，可替代上面的 Bash 重定向：

```powershell
Get-Content -Raw examples/bulk-create.json | flashdata bulk create --input - --run-id bulk-manual-001
Get-Content -Raw examples/bulk-update.json | flashdata bulk update bulk_YOUR_ID --input -
```

可选控制命令应单独执行：任务运行中才能暂停，检查状态后仅在 paused 时恢复；不要在上面已完成的流程之后直接执行：

```bash
flashdata bulk pause bulk_YOUR_ID
flashdata bulk get bulk_YOUR_ID
flashdata bulk resume bulk_YOUR_ID
```

签名上传能力仅保存在受保护的凭据库中，上传恢复绑定原始文件哈希。文件大小 1 字节至 2 GiB，每行一个视频 ID；签名上传有效期 15 分钟，内部结果保留 24 小时。能力过期需明确检查或联系支持，不自动创建替代任务。详见 [Bulk 指南](docs/bulk.md)。

部分 Bulk 结果保留已有回执并返回非零退出码，错误及导出信息输出到 stderr；从历史导出部分结果也不会被标记为成功。安全恢复读取可补全结果，并在核实任务已启动后清理已用过的签名上传能力。

## 配置、安全和诊断

全局选项包括 `--profile`、`--state-dir`、`--credential-store auto|os|file`、`--json`、`--non-interactive`、`--timeout`、`--help`。高级服务来源参数只配置新 profile，不能将旧凭据重定向到其他来源。默认状态目录位于仓库外的用户系统目录。Profile 使用小写字母、数字、点、下划线和连字符，以字母/数字开头。

Windows 默认使用当前用户 DPAPI；macOS 通过无 CGo 的原生 Keychain 调用；Linux 使用已解锁的 Secret Service，不请求 GUI 解锁，没有会话总线时新配置可自动使用私有文件模式。现有 OS 凭据库错误不会导致降级。显式 `--credential-store file` 用于无头持久环境；它是权限保护的明文封装，**不是加密**。详见[安全存储](docs/security.md)。

```bash
flashdata doctor
flashdata doctor --offline
flashdata account balance
flashdata account pricing --source google_search --request-count 100
flashdata account limits
flashdata account usage --source youtube_metadata --from 2026-10-01T00:00:00Z --to 2026-10-02T00:00:00Z
flashdata commands --json
flashdata auth logout
# Existing script state is migrated only through an explicit operation:
flashdata auth migrate --from /private/legacy/connection.json --profile migrated
```

Logout 清除本地凭据并保留历史，同时明确没有撤销远程访问；环境密钥需单独取消。迁移只通过显式命令进行，保留原文件、历史与防重投墓碑，不猜测旧 run ID、不重置损坏状态。

## 开发与验证

```bash
go test ./...
go vet ./...
# Supported hosts with a C compiler:
go test -race -short ./...
# Local archives and checksums; this script does not upload:
python scripts/build_release.py --go /path/to/go
```

离线测试使用本地模拟 API 和虚构密钥，覆盖编译后二进制、实际全新进程密钥验证、进程崩溃后锁释放、撤销后拒绝缓存、不确定提交、输入规范化和 Bulk 部分结果。不会使用真实密钥、不创建真实试用、不产生付费查询、不写云存储、不发布或部署。[验证记录](docs/validation.md) 列出实际检查及平台限制。

[命令参考](docs/commands.md) · [安全说明](SECURITY.md) · [贡献指南](CONTRIBUTING.md) · [更新记录](CHANGELOG.md) · [MIT 许可证](LICENSE)

[本地审查记录](docs/review.md)列出与 CLI 规范、后端契约逐项核对后修复的问题。

[联合发布检查](docs/release-readiness.md) · [第三方声明](THIRD_PARTY_NOTICES.md)
