# FlashData CLI v1.0.0

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs) · [English README](https://github.com/flashdatadev/flashdata-cli/blob/main/README.md) · [简体中文](https://github.com/flashdatadev/flashdata-cli/blob/main/README.zh-CN.md)

Initial standalone FlashData CLI, built on the pinned and checksum-verified **FlashData Go SDK v1.0.0**. Ready-to-run binaries are available from the [official v1.0.0 release](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.0).

## Included

- Secure named profiles with Windows DPAPI, macOS Keychain/Linux Secret Service integration, or an explicitly selected protected file store. Keys are entered through hidden terminal input or stdin, never command arguments.
- Recoverable machine onboarding, browser claiming and fresh-process credential verification before installation acknowledgement.
- Google search, YouTube data and media delivery; general query, background and mixed-source batch commands.
- Complete bulk create/upload/start/control/export workflows, streaming output, durable run IDs and recovery without automatic paid replay.
- Account/source discovery, diagnostics, legacy state migration, English and Chinese READMEs, runnable inputs and offline tests.

## Packages

| Platform | Archive |
| --- | --- |
| Linux amd64 | `flashdata_1.0.0_linux_amd64.tar.gz` |
| Linux arm64 | `flashdata_1.0.0_linux_arm64.tar.gz` |
| macOS Intel | `flashdata_1.0.0_darwin_amd64.tar.gz` |
| macOS Apple Silicon | `flashdata_1.0.0_darwin_arm64.tar.gz` |
| Windows amd64 | `flashdata_1.0.0_windows_amd64.zip` |

Archives contain the executable, English/Chinese READMEs, docs, example inputs and dependency license notices. The executable requires no separate Go, Node.js, Python or Docker installation. `build-report.json` records the exact CLI source commit, SDK version/checksum, toolchain, runtime source fingerprints and package hashes. The executable's `--version` reports 1.0.0 and that same source commit. The verified executables are unchanged from the tagged build; `documentationCommit` records the finalized package documentation separately, and executable hashes are included in the report.

Download the matching public archive and SHA256SUMS; no GitHub sign-in is required. Follow the [English installation guide](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.md) or [简体中文安装指南](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.zh-CN.md). Put the extracted executable in a user-controlled PATH directory (`flashdata.exe` on Windows), then run `flashdata --version` and `flashdata --help` offline. The READMEs also explain building directly from the v1.0.0 source tag with Go 1.25+.

## Checksums and verification

`SHA256SUMS` covers all five archives, `build-report.json`, the separately attached English/Chinese READMEs and these release notes. From a directory containing all those files:

```bash
sha256sum -c SHA256SUMS
```

PowerShell: use `Get-FileHash -Algorithm SHA256 <file>` and compare its hash with the corresponding `SHA256SUMS` entry. Packages are not signed or notarized; SHA-256 checksums verify bytes and are not publisher signatures.

## Validation and operational limits

CI runs full tests/vet on Linux, Windows and Intel macOS with Go 1.25/1.27, Linux race checks, Apple Silicon cross-builds on both Go versions and all five archive builds. Native Apple Silicon execution and live macOS Keychain/Linux Secret Service integration remain separate acceptance checks; no native ARM runtime result is claimed.

Onboarding is gated by server configuration. Imported-key recurring use requires the reviewed backend identity endpoint; unsupported identity responses fail closed. Publishing this CLI does not deploy backend changes, submit production paid work or verify customer storage. Lost or malformed submission responses remain uncertain: preserve the run ID and use read-only reconciliation/recovery before intentional new work. Cancellation stops local waiting while accepted remote work can continue.

## 简体中文摘要

独立 `flashdata` 命令行工具，锁定 Go SDK v1.0.0，覆盖安全凭据配置、可恢复接入、Google/YouTube 查询、后台批量任务、完整 Bulk 流程与持久化执行记录。提供 Linux amd64/arm64、macOS Intel/Apple Silicon、Windows amd64 安装包、构建记录和 SHA-256 校验和。GitHub Release 正式安装包现已提供公开下载；源码标签保持不变。原生 macOS 自动测试使用 Intel，Apple Silicon 已交叉编译，原生 ARM 运行及真实 Keychain/Secret Service 仍需单独验收。本次不部署后端或执行真实付费任务。
