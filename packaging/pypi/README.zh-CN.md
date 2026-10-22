# 通过 Python 包安装器使用 FlashData CLI

[English](README.md) | 简体中文

[FlashData](https://flashdata.dev) 提供结构化 Google 搜索、YouTube 数据、转录/字幕、后台任务和批量交付。本包安装官方 [FlashData CLI](https://github.com/flashdatadev/flashdata-cli)，支持安全配置、初始化和可恢复批量任务。命令名为 `flashdata`；Python SDK 是独立的 `flashdata` 发行包。

**状态：本地候选包，尚未发布至 PyPI。** 版本 **1.0.0** 封装已核验的现有 v1.0.0 可执行文件。发布前请使用下方本地 wheel 命令。轻量启动器需要 Python 3.10+，无需 Go 编译器或运行时，没有 Python 运行时依赖，安装和启动时均不下载可执行文件。

提供以下 wheel：Linux glibc 2.17+ x86_64/aarch64、macOS 13+ Intel/Apple Silicon、Windows x86_64。不提供 Linux musl/Alpine、Windows ARM64 或 32 位 wheel。不支持的平台会解析失败，不会安装其他架构的程序。

## 安装本地 wheel

先构建 `dist/wheels`，再从 CLI 仓库根目录执行。任选一种方式：pip/uv pip 使用独立 Python 环境，`uv tool` 管理独立 CLI 环境。安装后的命令仍为 `flashdata`。

```bash
python -m venv .venv-cli
```

Bash 通过 `source .venv-cli/bin/activate` 激活，PowerShell 使用 `.venv-cli\Scripts\Activate.ps1`。然后任选 pip 或 uv pip：

```bash
python -m pip install --no-index --find-links dist/wheels flashdata-cli==1.0.0
```

```bash
uv pip install --no-index --find-links dist/wheels flashdata-cli==1.0.0
```

也可在虚拟环境之外使用：

```bash
uv tool install --no-index --find-links dist/wheels flashdata-cli==1.0.0
flashdata --version
flashdata --help
```

若 uv 提示可执行文件目录不在 PATH，请运行 `uv tool update-shell` 并重新打开终端。旧的 Homebrew/APT/RPM/手动安装可能遮蔽新命令；Bash 用 `command -v flashdata`，PowerShell 用 `Get-Command flashdata` 检查。避免多个管理器同时管理同一个 PATH 条目。

今后发布到 PyPI 后，可使用 `pip install flashdata-cli`、`uv pip install flashdata-cli` 或 `uv tool install flashdata-cli`。**目前它们尚不是官方的在线安装方式。** 现在可以使用上方带 `--no-index --find-links` 的本地安装命令。

## 使用与维护

从 [FlashData 账户](https://flashdata.dev) 获取密钥，再运行 `flashdata auth login`，终端输入隐藏。参阅[英文初始化与安装](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.md)、[简体中文](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.zh-CN.md)及[完整批量流程](https://github.com/flashdatadev/flashdata-cli/blob/main/README.zh-CN.md)。启动器传递原生参数、标准输入/输出、信号和退出状态；安装命令不应包含 API 凭据。

获得新的本地 wheel 后，pip/uv pip 保留相同的 `--no-index --find-links` 选项并加上 `--upgrade`；uv tool 使用 `uv tool install --reinstall`。`==1.0.0` 会固定版本，需要升级时明确修改约束。按原安装方式分别使用 `python -m pip uninstall flashdata-cli`、`uv pip uninstall flashdata-cli` 或 `uv tool uninstall flashdata-cli` 卸载。卸载保留用户配置及运行历史；如需移除保存的凭据，请先执行 `flashdata auth logout`。

构建、验证及发布前提见仓库中的 `docs/python-packaging.zh-CN.md`。打包和测试脚本均不上传软件包。支持：[support@flashdata.dev](mailto:support@flashdata.dev)。
