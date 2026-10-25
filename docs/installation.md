# Install FlashData CLI v1.2.1

[FlashData](https://flashdata.dev) · [API docs](https://flashdata.dev/docs) · [简体中文](installation.zh-CN.md) · [Release downloads](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.2.1)

Choose an installation method below. Standalone archives and OS packages need no Go, Node.js, Python or Docker; pip and uv installation requires Python 3.10+. Downloading and offline version/help checks need no API key.

## Recommended setup

```sh
uv tool install flashdata-cli
flashdata setup
```

Setup selects your agent, installs the bundled skills and verifies access without a paid query. You can try it before signing up: a new profile without credentials requests temporary access, subject to trial availability and eligibility. Existing access is reused. See [onboarding](onboarding.md) for trial limits and signing up later, or [agent setup](agent-setup.md) for installation options. For an existing uv installation, run `uv tool upgrade flashdata-cli`, then rerun setup.

## Python (pip and uv)

Install [`flashdata-cli` from PyPI](https://pypi.org/project/flashdata-cli/) with Python 3.10+. It includes the official CLI executable; no Go installation is needed. Supported platforms are Linux glibc 2.17+ x86_64/ARM64, macOS 13+ Intel/Apple Silicon, and Windows x64.

Choose one method. Use a virtual environment for pip or uv pip; use `uv tool` for an independent CLI environment. The command is `flashdata`; the Python SDK is the separate `flashdata` package.

```bash
python -m venv .venv-cli
```

Activate it with `source .venv-cli/bin/activate` on Bash, or `.venv-cli\Scripts\Activate.ps1` on PowerShell. Then choose pip or uv pip:

```bash
python -m pip install --upgrade flashdata-cli
```

```bash
uv pip install --upgrade flashdata-cli
```

Alternatively, outside the virtual environment:

```bash
uv tool install flashdata-cli
flashdata --version
flashdata --help
```

If uv reports that its executable directory is not on PATH, run `uv tool update-shell` and restart the shell. A previous Homebrew/APT/RPM/manual installation can shadow this command; check `command -v flashdata` on Bash or `Get-Command flashdata` on PowerShell. Avoid using multiple managers for the same PATH entry.

Upgrade with the same pip or uv pip command above, or `uv tool upgrade flashdata-cli`. To pin a version, use `flashdata-cli==1.2.1`. Uninstall with `python -m pip uninstall flashdata-cli`, `uv pip uninstall flashdata-cli`, or `uv tool uninstall flashdata-cli`, matching your installer. Removal preserves profiles and run history; use `flashdata auth logout` first to remove a saved credential. For offline installs or local builds, see the [wheel packaging guide](python-packaging.md).

## APT (Ubuntu and Debian)

Use the [official signed APT repository](https://github.com/flashdatadev/apt) on Ubuntu 22.04/24.04 or Debian 12/13, with amd64 or arm64. Package **1.2.1-1** contains the same CLI **1.2.1** executable as the release archives; `-1` is the Debian packaging revision. No Go, Python or Node.js runtime is required.

Run this one-time Bash setup with sudo access. It checks the public key fingerprint, uses a repository-scoped `Signed-By` keyring and permits only the `flashdata` package from this source:

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

The dedicated certificate fingerprint is `19F337AC01C5F5B646A5583A04C7934A11D38FFA` (`FlashData APT Signing <support@flashdata.dev>`). Verify it against the [public key record](https://github.com/flashdatadev/apt/tree/main/keys). Do not disable signature/expiration checking or add the key to APT's global trusted keyring.

Upgrade after a newer CLI or packaging revision is published:

```bash
sudo apt-get update
sudo apt-get install --only-upgrade flashdata
flashdata --version
```

Inspect the package version and included examples/docs:

```bash
dpkg-query -W -f='${Version}\n' flashdata
apt-cache policy flashdata
ls /usr/share/doc/flashdata/examples
ls /usr/share/doc/flashdata/docs
man flashdata
```

To uninstall:

```bash
sudo apt-get remove flashdata
```

Removal or purge does not revoke keys or remove per-user profiles/run history. To remove a saved credential, run `flashdata auth logout` for that profile first. To also remove the repository configuration:

```bash
sudo rm -f /etc/apt/sources.list.d/flashdata.sources /etc/apt/preferences.d/flashdata /etc/apt/keyrings/flashdata.asc
sudo apt-get update
```

When switching from Homebrew or a manual archive, check `command -v flashdata`; an old executable earlier in PATH may hide `/usr/bin/flashdata`. Uninstall/move that old executable or adjust PATH, then restart the shell or run `hash -r`. Preserve your CLI state directory. [APT maintenance and validation](https://github.com/flashdatadev/apt/blob/main/docs/maintaining.md) explains signature freshness and upgrade testing.

## RPM (Fedora and RHEL)

Use the [official signed RPM repository](https://github.com/flashdatadev/rpm) for Fedora 43/44 and RHEL 8/9/10 userspace on x86_64/aarch64. Package **1.2.1-1** contains the exact official CLI **1.2.1** executable, docs, examples, man page and licenses, with no Go, Python or Node.js runtime dependency. Native tests cover Fedora and Red Hat UBI containers on both architectures; full RHEL hosts, subscription repositories and FIPS mode are not separately certified.

Run this one-time setup in Bash with sudo access. It verifies the full certificate fingerprint, imports the verified certificate and enables both metadata and package signature checks:

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

The dedicated RSA-4096 certificate fingerprint is `23AB58683E85FE7236D369713D19EA52D181CC36` (`FlashData RPM Signing <support@flashdata.dev>`), separate from APT. Compare with the [public key record](https://github.com/flashdatadev/rpm/tree/main/keys). DNF can request a second import for metadata: its per-repository metadata keyring is separate from the system-wide RPM package key database. RPM certificates are not scoped to one repository. `includepkgs=flashdata` limits packages from this source; keep `gpgcheck=1`, `repo_gpgcheck=1` and HTTPS verification enabled. `metadata_expire` controls cache freshness, not cryptographic expiry.

Upgrade when a newer CLI or RPM packaging release is available:

```bash
sudo dnf --refresh upgrade flashdata
rpm -q flashdata
/usr/bin/flashdata --version
```

Inspect package contents or uninstall:

```bash
rpm -ql flashdata
ls /usr/share/doc/flashdata/examples
man flashdata
sudo dnf remove flashdata
```

Removal preserves profiles, credentials and run history. Use `flashdata auth logout` for the selected profile first if you want to remove its saved credential. Remove `/etc/yum.repos.d/flashdata.repo` to stop using the source; removing its local key file does not remove the imported certificate from RPM's global key database. An administrator must check other uses before removing that database entry.

When switching from a manual or Homebrew installation, check `command -v flashdata`; an earlier PATH entry can hide `/usr/bin/flashdata`. Remove/move the previous executable or adjust PATH, then restart the shell or run `hash -r`. Keep the state directory. [RPM maintenance and validation](https://github.com/flashdatadev/rpm/blob/main/docs/maintaining.md) describes signed upgrade fixtures and key recovery.

## Homebrew (Linux and macOS)

Install [Homebrew](https://brew.sh) and follow its PATH instructions, then use the [official FlashData tap](https://github.com/flashdatadev/homebrew-tap):

```bash
brew install flashdatadev/tap/flashdata
flashdata --version
flashdata --help
```

Homebrew selects the matching Linux amd64/arm64 or macOS Intel/Apple Silicon archive and verifies its pinned SHA-256 checksum. It installs the standalone executable with no Go, Python or Node.js dependency. Installation and the offline checks above need no API key.

Upgrade when a newer version is published:

```bash
brew update
brew upgrade flashdatadev/tap/flashdata
```

The package also includes examples, English/Chinese release docs and dependency licenses:

```bash
ls "$(brew --prefix flashdata)/share/flashdata/examples"
ls "$(brew --prefix flashdata)/share/doc/flashdata"
```

To uninstall the package and optionally remove the tap:

```bash
brew uninstall flashdata
brew untap flashdatadev/tap
```

Uninstalling does not revoke API keys or delete saved profiles and run history. Use `flashdata auth logout` for a profile before uninstalling if you want to remove its saved credential. See [credential storage](security.md).

If switching from a manual installation, check `command -v flashdata`. A copy in `~/.local/bin` or another PATH directory may precede Homebrew. Move or remove that old executable, or adjust PATH; then restart the shell or run `hash -r`. Keep your existing CLI state directory so profiles and run recovery remain available.

## Manual archives (Linux and macOS)

The following Bash commands select Linux/macOS and amd64/arm64 automatically, download the official archive and SHA256SUMS over HTTPS, verify the selected archive before extraction, and install only the executable into `$HOME/.local/bin`. They require curl, tar, awk, install, and either sha256sum or shasum. Unsupported platforms and checksum errors stop installation.

```bash
(
  set -eu
  version=1.2.1
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

After a successful install, run `export PATH="$HOME/.local/bin:$PATH"` in your current terminal. Add `$HOME/.local/bin` to PATH in your shell startup configuration for future sessions. Downloaded docs and examples remain in a system temporary directory; current copies are also in the repository. Re-running installs this same version over the existing executable.

## Windows amd64 (PowerShell)

These commands verify the ZIP before extraction, install into `%LOCALAPPDATA%\FlashData\bin`, and update PATH for the current terminal:

```powershell
& {
    $ErrorActionPreference = 'Stop'
    $version = '1.2.1'
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

For future terminals, add `%LOCALAPPDATA%\FlashData\bin` to your user PATH in Windows Environment Variables. No administrator privileges are needed for this user-directory install.

## Next steps

After installing with any method, set up your agent and run your first search:

```bash
flashdata setup
flashdata search "battery recycling startups" --num 5 --json
```

No account or API key of your own is needed to request a trial. The default trial sources are Google Search and YouTube metadata, with a limited shared credit allowance. Use `flashdata auth status` to check access and `flashdata account balance` for remaining Credits. When ready to sign up, run `flashdata auth claim`, approve access in your browser, then run `flashdata auth finish --wait`. See [onboarding](onboarding.md) for details and other trial sources.

For terminal-only use, run `flashdata init --name "My CLI"` instead of setup; you do not need both. `setup --skip-auth` only installs skills offline and leaves credentials unchanged. Queries still need configured access.

Already have an API key? Use `flashdata auth login` and paste it at the hidden prompt. Automation can pipe a key from a secret manager to `auth login --api-key-stdin --non-interactive`. See [bulk workflows](bulk.md) and [run recovery](runs.md) for larger jobs.

On a fresh headless Linux setup, `flashdata auth login` automatically uses a private credential file if the desktop keyring is unavailable. Later commands reuse it without extra flags. The file has owner-only permissions and is not encrypted; see [credential storage](security.md).

The executable's `--version` reports 1.2.1 and the exact source commit recorded in `build-report.json`. Binaries and packaged documentation come from the same source tag. APT metadata and RPM packages/metadata are signed; standalone archives use SHA-256 checksums and macOS binaries are not notarized. Native archive acceptance runs on all five platforms; real Keychain/Secret Service behavior remains an environment-specific check. See the [release notes](release-notes-1.2.1.md).
