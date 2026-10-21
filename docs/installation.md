# Install FlashData CLI v1.0.0

[FlashData](https://flashdata.dev) · [API docs](https://flashdata.dev/docs) · [简体中文](installation.zh-CN.md) · [Release downloads](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.0)

Install the matching standalone executable. You do not need Go, Node.js, Python or Docker. Downloading and the offline version/help checks do not need an API key.

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

After a successful install, run `export PATH="$HOME/.local/bin:$PATH"` in your current terminal. Add `$HOME/.local/bin` to PATH in your shell startup configuration for future sessions. Downloaded docs and examples remain in a system temporary directory; current copies are also in the repository. Re-running installs this same version over the existing executable.

## Windows amd64 (PowerShell)

These commands verify the ZIP before extraction, install into `%LOCALAPPDATA%\FlashData\bin`, and update PATH for the current terminal:

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

For future terminals, add `%LOCALAPPDATA%\FlashData\bin` to your user PATH in Windows Environment Variables. No administrator privileges are needed for this user-directory install.

## Next steps

Run `flashdata auth login` to import an existing key from your [FlashData account](https://flashdata.dev). Terminal input is hidden. Automation can use a secret manager piped to `auth login --api-key-stdin --non-interactive`; never put a key in command arguments. See [onboarding](onboarding.md), [bulk workflows](bulk.md) and [run recovery](runs.md).

The executable's `--version` reports 1.0.0 and the source commit recorded in the release build report. Archives include updated docs; `documentationCommit` identifies that separate snapshot. Packages are not signed/notarized; checksums verify downloaded bytes. Native Apple Silicon execution and live Keychain/Secret Service acceptance remain separate checks, as described in the [release notes](release-notes-1.0.0.md).
