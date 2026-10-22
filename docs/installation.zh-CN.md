# 安装 FlashData CLI v1.0.1

[FlashData 官网](https://flashdata.dev) · [API 文档](https://flashdata.dev/docs) · [English](installation.md) · [正式版下载](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.1)

安装与平台匹配的独立可执行程序。独立归档和系统包无需 Go、Node.js、Python 或 Docker；Python 安装方式需要 Python 3.10+。下载及离线版本/帮助检查无需 API 密钥。

## Python 安装器（本地候选包）

**尚未发布到 PyPI。** 请按[打包指南](python-packaging.zh-CN.md)构建 `dist/wheels`。Wheel 需要 Python 3.10+，无需安装 Go，使用已核验的现有 CLI 二进制。平台范围、安装器隔离及维护说明见[软件包 README](../packaging/pypi/README.zh-CN.md)。

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

获得更新的本地 wheel 后，保留 `--no-index --find-links dist/wheels`，pip/uv pip 使用 `--upgrade`，uv tool 使用 `--reinstall`。明确修改固定版本约束；`==1.0.0` 会保持固定版本。分别通过 `python -m pip uninstall flashdata-cli`、`uv pip uninstall flashdata-cli` 或 `uv tool uninstall flashdata-cli` 卸载。卸载保留用户配置、凭据和执行历史；如需移除保存的凭据，请先执行 `flashdata auth logout`。

## APT（Ubuntu 和 Debian）

在 amd64 或 arm64 架构的 Ubuntu 22.04/24.04、Debian 12/13 上，可使用[官方签名 APT 软件源](https://github.com/flashdatadev/apt)。软件包 **1.0.1-1** 包含与发布归档一致的 CLI **1.0.1** 可执行文件；`-1` 是 Debian 打包修订号。无需 Go、Python 或 Node.js 运行时。

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

## RPM（Fedora 和 RHEL）

[官方签名 RPM 软件源](https://github.com/flashdatadev/rpm)支持 x86_64/aarch64 的 Fedora 43/44 及 RHEL 8/9/10 用户空间。软件包 **1.0.1-1** 包含与官方发布文件逐字节一致的 CLI **1.0.1**，以及文档、示例、手册页和许可证，无需 Go、Python 或 Node.js 运行时。原生测试覆盖两个架构的 Fedora 与 Red Hat UBI 容器；完整 RHEL 主机、订阅软件源和 FIPS 模式尚未单独认证。

使用 sudo 权限在 Bash 中执行一次以下配置。命令核对证书完整指纹、导入已验证证书，并同时启用元数据和软件包签名检查：

```bash
(
  set -eu
  case "$(rpm --eval '%{_arch}')" in
    x86_64|aarch64) ;;
    *) echo 'Supported RPM architectures: x86_64, aarch64' >&2; exit 1 ;;
  esac
  sudo dnf install -y ca-certificates gnupg2
  if ! command -v curl >/dev/null 2>&1; then sudo dnf install -y curl; fi
  work_dir="$(mktemp -d)"
  trap 'rm -rf -- "$work_dir"' EXIT
  curl -fsSL https://flashdatadev.github.io/rpm/flashdata.asc -o "$work_dir/flashdata.asc"
  expected=23AB58683E85FE7236D369713D19EA52D181CC36
  actual="$(gpg --batch --show-keys --with-colons "$work_dir/flashdata.asc" 2>/dev/null | awk -F: '$1 == "fpr" {print $10; exit}')"
  test "$actual" = "$expected" || { echo 'FlashData RPM key fingerprint mismatch' >&2; exit 1; }
  sudo install -d -m 0755 /etc/pki/rpm-gpg
  sudo install -m 0644 "$work_dir/flashdata.asc" /etc/pki/rpm-gpg/RPM-GPG-KEY-flashdata
  sudo rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-flashdata
  sudo tee /etc/yum.repos.d/flashdata.repo >/dev/null <<'EOF'
[flashdata]
name=FlashData CLI stable
baseurl=https://flashdatadev.github.io/rpm/stable/$basearch/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-flashdata
sslverify=1
includepkgs=flashdata
metadata_expire=6h
skip_if_unavailable=0
EOF
  sudo dnf -y --refresh install flashdata
  flashdata --version
  flashdata --help
)
```

独立的 RSA-4096 证书指纹为 `23AB58683E85FE7236D369713D19EA52D181CC36`，身份为 `FlashData RPM Signing <support@flashdata.dev>`，与 APT 证书分开管理。可对照[公开密钥记录](https://github.com/flashdatadev/rpm/tree/main/keys)核验。DNF 可能再次导入密钥用于元数据验证：每仓库独立的元数据密钥环与系统全局 RPM 软件包密钥库不同。RPM 证书不是仅对单一仓库生效；`includepkgs=flashdata` 限制本源提供的软件包。请保持 `gpgcheck=1`、`repo_gpgcheck=1` 和 HTTPS 校验；`metadata_expire` 仅控制缓存更新，不是密码学有效期。

新的 CLI 版本或 RPM 打包修订发布后，可运行：

```bash
sudo dnf --refresh upgrade flashdata
rpm -q flashdata
/usr/bin/flashdata --version
```

查看附带文件或卸载：

```bash
rpm -ql flashdata
ls /usr/share/doc/flashdata/examples
man flashdata
sudo dnf remove flashdata
```

卸载保留用户配置、凭据和执行历史。如需移除保存的凭据，请先为对应配置执行 `flashdata auth logout`。删除 `/etc/yum.repos.d/flashdata.repo` 可停用软件源；删除本地公钥文件不会移除 RPM 全局密钥库中的证书，管理员应确认其他用途后再移除该记录。

从手动安装或 Homebrew 迁移时，检查 `command -v flashdata`：PATH 中靠前的旧文件可能遮蔽 `/usr/bin/flashdata`。移走旧可执行文件或调整 PATH，再重启终端或执行 `hash -r`，并保留状态目录。[RPM 维护与验证](https://github.com/flashdatadev/rpm/blob/main/docs/maintaining.md)介绍了签名升级测试包及密钥恢复。

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
  version=1.0.1
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
    $version = '1.0.1'
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

`--version` 显示 1.0.1 及 `build-report.json` 记录的准确源码提交。可执行文件和附带文档来自同一个源码标签。APT 元数据以及 RPM 软件包/元数据均有签名；独立归档使用 SHA-256 校验，macOS 可执行文件未经过公证。五个平台均运行原生归档验收；真实 Keychain/Secret Service 行为仍需按具体环境检查。详见[版本说明](release-notes-1.0.1.md)。
