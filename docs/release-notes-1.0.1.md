# FlashData CLI v1.0.1

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs) · [English README](https://github.com/flashdatadev/flashdata-cli#readme) · [简体中文](https://github.com/flashdatadev/flashdata-cli/blob/main/README.zh-CN.md)

## What's new

- Public `flashdata discover plans` and `flashdata discover status` commands are now included in the standalone binaries. They read public API endpoints without an API key or account credentials.
- English and Chinese workflow selectors and examples cover research, AI context, news, YouTube search, metadata, existing automatic transcripts, uploaded captions and complete MP3/MP4 bulk downloads to customer storage.
- Clear API coverage and SDK comparisons explain accepted task recovery, partial item failures, export expiry and how to avoid duplicate submissions.
- Homebrew, signed APT and signed RPM installation and upgrade instructions point to v1.0.1. Existing profiles, credentials and run journals are preserved during package upgrades.

## Install or upgrade

No Go compiler, Python, Node.js or source build is required. See the [English installation guide](https://github.com/flashdatadev/flashdata-cli/blob/v1.0.1/docs/installation.md) or [简体中文安装指南](https://github.com/flashdatadev/flashdata-cli/blob/v1.0.1/docs/installation.zh-CN.md) for first-time repository setup.

| Method | Install | Upgrade an existing installation |
| --- | --- | --- |
| Homebrew (Linux/macOS) | `brew install flashdatadev/tap/flashdata` | `brew update && brew upgrade flashdatadev/tap/flashdata` |
| APT (Ubuntu/Debian) | `sudo apt-get install flashdata` | `sudo apt-get update && sudo apt-get install --only-upgrade flashdata` |
| RPM (Fedora/RHEL) | `sudo dnf --refresh install flashdata` | `sudo dnf --refresh upgrade flashdata` |
| Manual archive | Download, verify and extract the matching archive | Replace only the executable; retain the CLI state directory |

APT/RPM commands require the official signed repository to be configured first. Their package version is `1.0.1-1`. Check `flashdata --version` after upgrading; it reports `1.0.1` and the exact tagged source commit. An older copy earlier in PATH can hide a package-manager installation.

## Downloads and verification

| Platform | Archive |
| --- | --- |
| Linux amd64 | `flashdata_1.0.1_linux_amd64.tar.gz` |
| Linux arm64 | `flashdata_1.0.1_linux_arm64.tar.gz` |
| macOS Intel | `flashdata_1.0.1_darwin_amd64.tar.gz` |
| macOS Apple Silicon | `flashdata_1.0.1_darwin_arm64.tar.gz` |
| Windows x64 | `flashdata_1.0.1_windows_amd64.zip` |

Download from the [official release](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.1) and verify the selected archive against `SHA256SUMS` before extraction. Archives include the executable, both READMEs, docs, example inputs and dependency licenses from the same source tag. `build-report.json` records the CLI commit, Go toolchain, pinned Go SDK v1.0.0 checksum and archive hashes. Previous v1.0.0 assets and tags remain unchanged.

Source tests and native archive acceptance cover all five platforms using mock APIs, including onboarding, credentials, query/bulk execution and recovery. Homebrew checks install and upgrades on Linux/macOS Intel and ARM64; signed repository checks cover Ubuntu 22.04/24.04, Debian 12/13, Fedora 43/44 and Red Hat UBI 8/9/10 userspace on both CPUs. Full RHEL hosts, FIPS mode, live OS credential services and production API/cloud delivery are separate environment-specific checks. Standalone archives use checksums; macOS binaries are not notarized. APT metadata and RPM packages/metadata use the existing signing certificates.

No PyPI, TestPyPI or npm publication is part of this release. The separately prepared Python wheel candidate remains pinned to v1.0.0 and does not contain the new discovery commands.

## 简体中文

v1.0.1 将无需 API 密钥的 `discover plans`、`discover status` 纳入五个平台的独立安装包，无需源码构建。中英文文档补充工作流选择、手动上传字幕、MP3/MP4 完整批量下载及恢复示例。Homebrew、签名 APT 和 RPM 源提供升级；升级保留用户配置、凭据和执行历史。首次安装请先按安装指南配置软件源，升级后用 `flashdata --version` 确认版本为 1.0.1。独立归档附带准确源码提交、构建记录和 SHA-256 校验和；旧 v1.0.0 文件保持不变。本次不发布 PyPI、TestPyPI 或 npm，独立 Python wheel 候选包仍固定为 v1.0.0。
