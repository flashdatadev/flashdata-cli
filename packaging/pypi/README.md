# FlashData CLI for Python package installers

English | [简体中文](README.zh-CN.md)

[FlashData](https://flashdata.dev) provides structured Google search and YouTube data, transcripts/captions, background jobs and bulk delivery. This package installs the official [FlashData CLI](https://github.com/flashdatadev/flashdata-cli), including secure profiles, onboarding and recoverable bulk workflows. The command is `flashdata`; the Python SDK is the separate `flashdata` distribution.

**Status: local release candidate, not published to PyPI.** Version **1.0.0** wraps the verified existing v1.0.0 executable. Use the local wheel instructions below until publication. The package requires Python 3.10+ for its small launcher and no Go compiler or runtime. It has no Python runtime dependencies and never downloads an executable during installation or startup.

Supported wheels: Linux glibc 2.17+ x86_64/aarch64, macOS 13+ Intel/Apple Silicon, Windows x86_64. Linux musl/Alpine, Windows ARM64 and 32-bit systems have no matching wheels. An unsupported platform fails resolution rather than installing a different executable.

## Install local wheels

Run from the CLI checkout after building `dist/wheels`. Choose one method. Use an isolated Python environment for pip/uv pip; use `uv tool` for an independent CLI environment. The command remains `flashdata`.

```bash
python -m venv .venv-cli
```

Activate it with `source .venv-cli/bin/activate` on Bash, or `.venv-cli\Scripts\Activate.ps1` on PowerShell. Then choose pip or uv pip:

```bash
python -m pip install --no-index --find-links dist/wheels flashdata-cli==1.0.0
```

```bash
uv pip install --no-index --find-links dist/wheels flashdata-cli==1.0.0
```

Alternatively, outside the virtual environment:

```bash
uv tool install --no-index --find-links dist/wheels flashdata-cli==1.0.0
flashdata --version
flashdata --help
```

If uv reports that its executable directory is not on PATH, run `uv tool update-shell` and restart the shell. A previous Homebrew/APT/RPM/manual installation can shadow this command; check `command -v flashdata` on Bash or `Get-Command flashdata` on PowerShell. Avoid using multiple managers for the same PATH entry.

After a future PyPI publication, the registry commands will be `pip install flashdata-cli`, `uv pip install flashdata-cli`, or `uv tool install flashdata-cli`. **They are not currently an official registry installation method.** The local `--no-index --find-links` instructions above are usable now.

## Use and maintain

Get a key from your [FlashData account](https://flashdata.dev), then use `flashdata auth login` (hidden terminal input). See [English onboarding and installation](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.md), [简体中文](https://github.com/flashdatadev/flashdata-cli/blob/main/docs/installation.zh-CN.md) and the [complete bulk workflow](https://github.com/flashdatadev/flashdata-cli#complete-bulk-workflow). Native arguments, stdin/stdout/stderr, signals and exit status pass through the launcher; no API credentials belong in installation commands.

For an updated local wheel set, use the same `--no-index --find-links` options and `--upgrade` with pip/uv pip, or `uv tool install --reinstall` with uv tool. An exact `==1.0.0` constraint stays pinned; change it intentionally for a newer version. Uninstall with `python -m pip uninstall flashdata-cli`, `uv pip uninstall flashdata-cli`, or `uv tool uninstall flashdata-cli`, matching the installer you used. Removing the package preserves per-user CLI profiles and run history; use `flashdata auth logout` first if you want to remove a saved credential.

Build, validation and publication prerequisites are in the checkout's `docs/python-packaging.md`. Nothing in the packaging or test scripts uploads packages. Support: [support@flashdata.dev](mailto:support@flashdata.dev).
