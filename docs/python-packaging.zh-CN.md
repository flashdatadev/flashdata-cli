# Python wheel 打包

[FlashData](https://flashdata.dev) · [English](python-packaging.md) · [安装指南](installation.zh-CN.md#pythonpip-和-uv)

[`flashdata-cli` 1.1.0 已发布至 PyPI](https://pypi.org/project/flashdata-cli/)。通过 pip 或 uv 安装请参阅[安装指南](installation.zh-CN.md#pythonpip-和-uv)。本指南介绍本地 wheel 的构建与验证。软件包提供 `flashdata` 命令及 `python -m flashdata_cli`；Python SDK 是独立的 `flashdata` 发行包。

## 兼容性与内容

无依赖的 Python 3.10+ 启动器执行内置的官方 v1.1.0 二进制。Linux/macOS 用 `execv` 保留 PID 和原生信号处理；Windows 等待共享控制台的子进程。参数不经 shell，继承标准输入/输出和环境，并保留退出状态。启动器不加载 `.env`、读取凭据、下载二进制或管理配置；认证、存储和执行恢复仍由原生 CLI 负责。

| Wheel 平台 | 目标 | 最低要求 |
| --- | --- | --- |
| `manylinux_2_17_x86_64.manylinux2014_x86_64` | Linux amd64 | glibc 2.17 |
| `manylinux_2_17_aarch64.manylinux2014_aarch64` | Linux arm64 | glibc 2.17 |
| `macosx_13_0_x86_64` | macOS Intel | macOS 13 |
| `macosx_13_0_arm64` | macOS Apple Silicon | macOS 13 |
| `win_amd64` | Windows amd64 | 上游支持的 64 位 Windows |

`py3-none` 不依赖 CPython ABI；`Requires-Python: >=3.10` 指定解释器下限。Linux 二进制为静态程序，不含 ELF `PT_INTERP`；macOS 下限来自真实 Mach-O 部署目标。不声明 musllinux/Alpine、Windows ARM64、32 位或通用 `any` wheel。不支持的平台会解析失败，安装时不会下载或调用 Go 编译器。

Wheel 在 `flashdata_cli/share` 中包含上游中英文文档、示例及依赖许可证，在 `.dist-info/licenses` 中包含项目许可证，并记录归档/可执行文件哈希和原生源码提交。没有 Python 运行时依赖，安装/启动时不下载内容。本候选包仅提供 wheel，不支持可编辑安装。

## 构建

从仓库根目录执行，创建环境后，Bash 使用 `source .venv-wheel-build/bin/activate` 激活，PowerShell 使用 `.venv-wheel-build\Scripts\Activate.ps1`：

```bash
python -m venv .venv-wheel-build
```

```bash
python -m pip install -r packaging/pypi/requirements-build.txt
python scripts/build_wheels.py --download
```

`--download` 明确允许从官方 GitHub Release 下载缺少的归档，缓存于 `.cache/wheel-archives`，使用前对照 `packaging/pypi/upstream.json` 核验。已有归档时可离线构建：

```bash
python scripts/build_wheels.py --archive-dir /path/to/verified-release-archives --output dist/wheels-offline
```

一台主机可构建全部五个平台；`--target` 可选择单独的 `linux/amd64`、`linux/arm64`、`darwin/amd64`、`darwin/arm64` 或 `windows/amd64`。输出目录必须为空，防止覆盖已审阅的候选包。构建保留原生二进制字节，即使在 Windows 上也明确记录 Unix 执行权限，并生成可复现 wheel、`build-report.json` 和 `SHA256SUMS`。

标准 PEP 517 项目位于 `packaging/pypi/pyproject.toml`，使用固定版本的 Hatchling。高级构建可先设置 `FLASHDATA_CLI_ARCHIVE_DIR`，再运行 `python -m build --wheel packaging/pypi`；可选 `FLASHDATA_WHEEL_TARGET` 选择目标，`FLASHDATA_WHEEL_DOWNLOAD=1` 明确允许下载。常规构建脚本负责管理这些设置并验证结果。本候选版本不包含源码发行包。

## 验证

```bash
python -m unittest discover -s scripts -p 'test_*.py'
python -m twine check --strict dist/wheels/*.whl
python scripts/test_wheel_resolution.py
python scripts/test_wheel_install.py --uv uv
```

安装测试隔离 pip/uv pip 环境和 uv 工具目录，禁用索引及 uv 网络访问，检查版本/帮助/命令，对比原生和封装程序错误输出及退出状态，核验二进制哈希与文档/许可证，执行重装/卸载并检查用户状态标记保留。不会修改日常 uv 工具或 PATH，也不调用付费 API。重复运行请指定新的 `--work-dir`。

测试另一份已安装的解释器：

```bash
python scripts/test_wheel_install.py --uv uv --python /path/to/python3.10 --work-dir .cache/wheel-install-python310
```

`scripts/wheel_validation.py` 独立检查平台标签、二进制格式/最低系统版本、元数据、许可证/文档以及全部 `RECORD` 哈希。重复构建至另一空目录应得到相同哈希。`wheels.yml` 构建候选包，并在五种系统/CPU 和 Python 3.10/3.15 上执行原生检查。[GitHub Actions 结果](https://github.com/flashdatadev/flashdata-cli/actions/workflows/wheels.yml)与候选产物独立于 PyPI 正式发布。

[v1.1.0 原生验证任务](https://github.com/flashdatadev/flashdata-cli/actions/runs/39943624896)通过全部 **30 种安装组合**：五种原生平台，各使用 Python **3.10 和 3.15**，分别运行 pip、uv pip 和 uv tool。每种组合验证安装、离线版本/帮助/命令、参数/标准输入/错误/退出状态传递、原生二进制哈希、附带文档/许可证、重装、卸载及用户状态保留。[CLI 检查](https://github.com/flashdatadev/flashdata-cli/actions/runs/39943624704)也全部通过。

| 原生平台 | Python 版本 | pip / uv pip / uv tool |
| --- | --- | --- |
| Linux x86_64 | 3.10, 3.15 | 通过 |
| Linux aarch64 | 3.10, 3.15 | 通过 |
| macOS Intel | 3.10, 3.15 | 通过 |
| macOS Apple Silicon | 3.10, 3.15 | 通过 |
| Windows x86_64 | 3.10, 3.15 | 通过 |

五种 v1.1.0 wheel 均通过严格的元数据/完整性检查，重复构建哈希一致。pip 能选择支持的平台标签，拒绝不支持的平台和 Python 3.9。CI 测试对象是候选 wheel，不代表已认证线上服务可用性或真实系统凭据库行为。

无需 PyPI 即可试用本地候选包：下载成功工作流中的 `wheel-candidates` 产物并解压至仓库的 `dist/wheels`。在虚拟环境中运行 `python -m pip install --no-index --find-links dist/wheels flashdata-cli==1.1.0`，或使用带相同选项的 `uv pip install`。CI 产物保留七天；过期后可从固定版本归档重新构建。

## 发布前提

确认原生 CI 矩阵成功，审核内容/哈希，确认官方 PyPI 账户/项目所有权，再单独配置经明确授权的发布流程。不要更改已有 GitHub Release 标签/产物。分别记录打包提交和原生源码提交；正式发布后再同步更新两种语言的候选提示及在线安装命令。

准备工作流只有仓库读取权限，没有注册表上传步骤或发布凭据。升级原生版本时，先发布并测试 CLI，再更新项目版本、固定归档/哈希、真实平台下限和文档。PyPI 发布后的纯打包修复需要新的 PEP 440 版本，例如 `1.0.0.post1`，不得覆盖已上传的 wheel。当前构建有意要求 Python 包和原生版本一致；若引入 post 版本，必须明确调整此检查并保留原生来源信息。

支持：[support@flashdata.dev](mailto:support@flashdata.dev)。
