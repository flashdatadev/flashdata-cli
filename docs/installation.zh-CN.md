# 安装 FlashData CLI v1.0.0

[FlashData 官网](https://flashdata.dev) · [API 文档](https://flashdata.dev/docs) · [English](installation.md) · [正式版下载](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.0)

安装与平台匹配的独立可执行程序，无需 Go、Node.js、Python 或 Docker。下载以及离线版本/帮助检查均无需 API 密钥。

## APT（Ubuntu 和 Debian）

在 amd64 或 arm64 架构的 Ubuntu 22.04/24.04、Debian 12/13 上，可使用[官方签名 APT 软件源](https://github.com/flashdatadev/apt)。软件包 **1.0.0-1** 包含与发布归档一致的 CLI **1.0.0** 可执行文件；`-1` 是 Debian 打包修订号。无需 Go、Python 或 Node.js 运行时。

使用 sudo 权限在 Bash 中执行一次以下配置。命令会核对公钥指纹，通过专用 `Signed-By` 密钥环限制信任范围，并只允许从该软件源选择 `flashdata` 软件包：

```bash
(
  set -eu
  arch="$(dpkg --print-architecture)"
  case "$arch" in amd64|arm64) ;; *) echo "Unsupported architecture: $arch" >&2; exit 1 ;; esac
  sudo apt-get update
  sudo apt-get install -y ca-certificates curl gnupg
  work_dir="$(mktemp -d)"
  trap 'rm -rf -- "$work_dir"' EXIT
  curl --fail --location https://flashdatadev.github.io/apt/flashdata.asc --output "$work_dir/flashdata.asc"
  expected=19F337AC01C5F5B646A5583A04C7934A11D38FFA
  actual="$(gpg --batch --show-keys --with-colons "$work_dir/flashdata.asc" 2>/dev/null | awk -F: '$1 == "fpr" { print $10; exit }')"
  test "$actual" = "$expected"
  sudo install -d -m 0755 /etc/apt/keyrings
  sudo install -m 0644 "$work_dir/flashdata.asc" /etc/apt/keyrings/flashdata.asc
  sudo tee /etc/apt/sources.list.d/flashdata.sources >/dev/null <<EOF
Types: deb
URIs: https://flashdatadev.github.io/apt/
Suites: stable
Components: main
Architectures: $arch
Signed-By: /etc/apt/keyrings/flashdata.asc $expected
EOF
  sudo tee /etc/apt/preferences.d/flashdata >/dev/null <<'EOF'
Package: flashdata
Pin: release o=FlashData
Pin-Priority: 500

Package: *
Pin: release o=FlashData
Pin-Priority: -1
EOF
  sudo apt-get update
  sudo apt-get install -y flashdata
  flashdata --version
  flashdata --help
)
```

专用证书指纹为 `19F337AC01C5F5B646A5583A04C7934A11D38FFA`，身份为 `FlashData APT Signing <support@flashdata.dev>`；可对照[公钥记录](https://github.com/flashdatadev/apt/tree/main/keys)核验。请勿禁用签名或有效期检查，也不要将此公钥添加到 APT 全局信任密钥环。

新的 CLI 版本或打包修订发布后，可运行以下命令升级：

```bash
sudo apt-get update
sudo apt-get install --only-upgrade flashdata
flashdata --version
```

查看软件包版本及附带示例/文档：

```bash
dpkg-query -W -f='${Version}\n' flashdata
apt-cache policy flashdata
ls /usr/share/doc/flashdata/examples
ls /usr/share/doc/flashdata/docs
man flashdata
```

卸载命令：

```bash
sudo apt-get remove flashdata
```

卸载或清除软件包不会撤销密钥，也不会删除用户配置或执行历史。如需移除保存的凭据，请先为对应配置运行 `flashdata auth logout`。如需同时移除软件源配置：

```bash
sudo rm -f /etc/apt/sources.list.d/flashdata.sources /etc/apt/preferences.d/flashdata /etc/apt/keyrings/flashdata.asc
sudo apt-get update
```

从 Homebrew 或手动归档安装迁移时，请检查 `command -v flashdata`：PATH 中靠前的旧版本可能遮蔽 `/usr/bin/flashdata`。卸载/移走旧可执行文件或调整 PATH，再重启终端或执行 `hash -r`；保留 CLI 状态目录。[APT 维护与验证说明](https://github.com/flashdatadev/apt/blob/main/docs/maintaining.md)介绍了签名有效期和升级测试。

## Homebrew（Linux 和 macOS）

先安装 [Homebrew](https://brew.sh) 并按提示配置 PATH，然后使用 [FlashData 官方软件源](https://github.com/flashdatadev/homebrew-tap)：

```bash
brew install flashdatadev/tap/flashdata
flashdata --version
flashdata --help
```

Homebrew 自动选择 Linux amd64/arm64 或 macOS Intel/Apple Silicon 归档，并验证固定的 SHA-256 校验和。安装的是独立可执行文件，无需 Go、Python 或 Node.js；安装及以上离线检查无需 API 密钥。

有新版本发布后，可运行以下命令升级：

```bash
brew update
brew upgrade flashdatadev/tap/flashdata
```

软件包还包含示例、中英文版本文档及依赖许可证：

```bash
ls "$(brew --prefix flashdata)/share/flashdata/examples"
ls "$(brew --prefix flashdata)/share/doc/flashdata"
```

卸载软件包并按需移除软件源：

```bash
brew uninstall flashdata
brew untap flashdatadev/tap
```

卸载不撤销 API 密钥，也不删除已保存的配置和执行历史。若要移除某个配置中已保存的凭据，请在卸载前为该配置运行 `flashdata auth logout`，详见[凭据存储](security.md)。

从手动安装迁移时，请先用 `command -v flashdata` 检查当前可执行文件。`~/.local/bin` 或其他 PATH 目录中的旧版本可能排在 Homebrew 前面；移走旧可执行文件或调整 PATH，再重启终端或执行 `hash -r`。请保留现有 CLI 状态目录，以继续使用配置和执行恢复。

## 手动安装归档（Linux 和 macOS）

以下 Bash 命令自动识别 Linux/macOS 与 amd64/arm64，从官方 HTTPS 地址下载归档和 SHA256SUMS，先验证所选归档，再解压，并仅将可执行程序安装到 `$HOME/.local/bin`。需要 curl、tar、awk、install，以及 sha256sum 或 shasum。不支持的平台或校验错误会终止安装。

```bash
(
  set -eu
  version=1.0.0
  case "$(uname -s)" in
    Linux) platform=linux ;;
    Darwin) platform=darwin ;;
    *) echo "Unsupported operating system" >&2; exit 1 ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64) architecture=amd64 ;;
    arm64|aarch64) architecture=arm64 ;;
    *) echo "Unsupported CPU architecture" >&2; exit 1 ;;
  esac
  archive="flashdata_${version}_${platform}_${architecture}.tar.gz"
  release_url="https://github.com/flashdatadev/flashdata-cli/releases/download/v${version}"
  work_dir="$(mktemp -d)"
  curl --fail --location "$release_url/$archive" --output "$work_dir/$archive"
  curl --fail --location "$release_url/SHA256SUMS" --output "$work_dir/SHA256SUMS"
  expected="$(awk -v file="$archive" '$2 == file { print $1 }' "$work_dir/SHA256SUMS")"
  test -n "$expected"
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$work_dir/$archive" | awk '{print $1}')"
  else
    actual="$(shasum -a 256 "$work_dir/$archive" | awk '{print $1}')"
  fi
  test "$actual" = "$expected"
  tar -xzf "$work_dir/$archive" -C "$work_dir"
  install -d "$HOME/.local/bin"
  install -m 755 "$work_dir/flashdata" "$HOME/.local/bin/flashdata"
  "$HOME/.local/bin/flashdata" --version
  "$HOME/.local/bin/flashdata" --help
)
```

安装成功后，在当前终端执行 `export PATH="$HOME/.local/bin:$PATH"`。请在 shell 启动配置中将 `$HOME/.local/bin` 加入 PATH，供后续终端使用。下载的文档与示例保留在系统 `mktemp` 生成的临时目录中；仓库也提供最新副本。重复运行会用同一版本覆盖现有可执行程序。

## Windows amd64（PowerShell）

以下命令先校验 ZIP 再解压，安装到 `%LOCALAPPDATA%\FlashData\bin`，并为当前终端更新 PATH：

```powershell
& {
    $ErrorActionPreference = 'Stop'
    $version = '1.0.0'
    $archive = "flashdata_${version}_windows_amd64.zip"
    $releaseUrl = "https://github.com/flashdatadev/flashdata-cli/releases/download/v$version"
    $workDir = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString())
    New-Item -ItemType Directory -Path $workDir | Out-Null
    $archivePath = Join-Path $workDir $archive
    $sumsPath = Join-Path $workDir 'SHA256SUMS'
    Invoke-WebRequest "$releaseUrl/$archive" -OutFile $archivePath -UseBasicParsing
    Invoke-WebRequest "$releaseUrl/SHA256SUMS" -OutFile $sumsPath -UseBasicParsing
    $entry = @(Get-Content $sumsPath | Where-Object { $_ -match ('  ' + [regex]::Escape($archive) + '$') })
    if ($entry.Count -ne 1) { throw 'Missing or duplicate checksum entry' }
    $expected = ($entry[0] -split '\s+')[0]
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $archivePath).Hash
    if ($actual -ne $expected) { throw 'SHA-256 checksum mismatch' }
    Expand-Archive -LiteralPath $archivePath -DestinationPath $workDir -Force
    $installDir = Join-Path $env:LOCALAPPDATA 'FlashData\bin'
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $workDir 'flashdata.exe') -Destination $installDir -Force
    $env:PATH = "$installDir;$env:PATH"
    flashdata --version
    if ($LASTEXITCODE -ne 0) { throw 'FlashData version check failed' }
    flashdata --help
    if ($LASTEXITCODE -ne 0) { throw 'FlashData help check failed' }
}
```

如需在后续终端使用，请在 Windows“环境变量”设置中将 `%LOCALAPPDATA%\FlashData\bin` 加入用户 PATH。安装到此用户目录不需要管理员权限。

## 下一步

运行 `flashdata auth login`，导入 [FlashData 账户](https://flashdata.dev)中的已有密钥，终端输入会隐藏。自动化可将密钥管理器的输出通过管道传给 `auth login --api-key-stdin --non-interactive`；不要将密钥放入命令行参数。详见[接入](onboarding.md)、[Bulk 流程](bulk.md)及[执行恢复](runs.md)。

`--version` 显示 1.0.0 及正式版构建记录中的源码提交。归档包含更新后的文档；`documentationCommit` 单独标识该文档快照。安装包未签名或公证，校验和用于验证下载字节。原生 Apple Silicon 运行及真实 Keychain/Secret Service 验收仍需单独检查，详见[版本说明](release-notes-1.0.0.md)。
