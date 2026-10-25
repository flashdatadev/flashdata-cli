# Python wheel packaging

[FlashData](https://flashdata.dev) · [简体中文](python-packaging.zh-CN.md) · [Installation](installation.md#python-pip-and-uv)

[`flashdata-cli` on PyPI](https://pypi.org/project/flashdata-cli/) installs the official FlashData CLI. For pip and uv installation, see the [installation guide](installation.md#python-pip-and-uv). This guide covers building and validating local wheels. The package provides `flashdata` and `python -m flashdata_cli`; the Python SDK is the separate `flashdata` distribution.

## Compatibility and contents

The dependency-free Python 3.10+ launcher runs the bundled official v1.2.0 executable. Linux/macOS use `execv` to preserve PID and native signal handling; Windows waits for its child sharing the console. Arguments pass without a shell, with inherited stdin/stdout/stderr and environment and preserved exit status. The launcher does not load `.env`, read credentials, download binaries or manage profiles. Authentication, storage and run recovery remain native CLI responsibilities.

| Wheel platform | Target | Minimum |
| --- | --- | --- |
| `manylinux_2_17_x86_64.manylinux2014_x86_64` | Linux amd64 | glibc 2.17 |
| `manylinux_2_17_aarch64.manylinux2014_aarch64` | Linux arm64 | glibc 2.17 |
| `macosx_13_0_x86_64` | macOS Intel | macOS 13 |
| `macosx_13_0_arm64` | macOS Apple Silicon | macOS 13 |
| `win_amd64` | Windows amd64 | Upstream-supported 64-bit Windows |

`py3-none` has no CPython ABI dependency; `Requires-Python: >=3.10` sets the interpreter minimum. Linux binaries are static without ELF `PT_INTERP`; macOS minima come from the actual Mach-O deployment target. No musllinux/Alpine, Windows ARM64, 32-bit or universal `any` wheel is advertised. Unsupported platforms fail resolution; no Go compiler is downloaded or invoked during installation.

Wheels include upstream English/Chinese docs, examples and dependency licenses under `flashdata_cli/share`, the project license in `.dist-info/licenses`, and provenance with archive/executable hashes and the native source commit. There are no Python runtime dependencies or install/startup downloads. This candidate contains wheels only; editable installs are rejected.

## Build

Run from the repository root. Create an environment, then activate it with `source .venv-wheel-build/bin/activate` on Bash or `.venv-wheel-build\Scripts\Activate.ps1` on PowerShell:

```bash
python -m venv .venv-wheel-build
```

```bash
python -m pip install -r packaging/pypi/requirements-build.txt
python scripts/build_wheels.py --download
```

`--download` explicitly fetches missing archives from the official GitHub release into `.cache/wheel-archives`, checking `packaging/pypi/upstream.json` before use. With archives available, build offline:

```bash
python scripts/build_wheels.py --archive-dir /path/to/verified-release-archives --output dist/wheels-offline
```

One host can build all five wheels; `--target` selects a single `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` or `windows/amd64` target. Output must be empty to avoid overwriting reviewed candidates. Builds preserve executable bytes, encode Unix executable permissions on Windows and produce reproducible wheels, `build-report.json` and `SHA256SUMS`.

The standard PEP 517 project is `packaging/pypi/pyproject.toml`, using pinned Hatchling. Advanced builds can set `FLASHDATA_CLI_ARCHIVE_DIR` and run `python -m build --wheel packaging/pypi`; optional `FLASHDATA_WHEEL_TARGET` chooses a target and `FLASHDATA_WHEEL_DOWNLOAD=1` explicitly enables downloads. The normal build script manages these settings and validates output. No source distribution is part of this candidate.

## Validate

```bash
python -m unittest discover -s scripts -p 'test_*.py'
python -m twine check --strict dist/wheels/*.whl
python scripts/test_wheel_resolution.py
python scripts/test_wheel_install.py --uv uv
```

The installer test isolates pip/uv pip environments and uv tool directories, disables indexes and uv network access, checks version/help/commands, compares native/wrapped error output and exit status, verifies executable hashes and docs/licenses, reinstalls/uninstalls, and checks a preserved user-state marker. It leaves your normal uv tools and PATH unchanged and makes no paid API calls. Repeat with a new `--work-dir`.

Test another installed interpreter:

```bash
python scripts/test_wheel_install.py --uv uv --python /path/to/python3.10 --work-dir .cache/wheel-install-python310
```

`scripts/wheel_validation.py` independently checks tags, executable format/minimum OS, metadata, licenses/docs and every `RECORD` hash. Rebuilding into another empty output should produce identical wheel hashes. The `wheels.yml` workflow builds candidates and runs native tests on five OS/CPU combinations and Python 3.10/3.15. Its [GitHub Actions results](https://github.com/flashdatadev/flashdata-cli/actions/workflows/wheels.yml) and candidate artifacts are separate from a PyPI release.

The [v1.1.0 native validation run](https://github.com/flashdatadev/flashdata-cli/actions/runs/39943624896) passed all **30 installer combinations**: pip, uv pip and uv tool on five native platforms, each with Python **3.10 and 3.15**. Each combination verified installation, offline version/help/commands, argument/stdin/error/exit forwarding, exact executable hashes, bundled docs/licenses, reinstall, uninstall and preserved user state. The [CLI checks](https://github.com/flashdatadev/flashdata-cli/actions/runs/39943624704) also passed.

| Native platform | Python versions | pip / uv pip / uv tool |
| --- | --- | --- |
| Linux x86_64 | 3.10, 3.15 | Passed |
| Linux aarch64 | 3.10, 3.15 | Passed |
| macOS Intel | 3.10, 3.15 | Passed |
| macOS Apple Silicon | 3.10, 3.15 | Passed |
| Windows x86_64 | 3.10, 3.15 | Passed |

All five v1.1.0 wheels passed strict metadata/integrity checks and reproduced with identical hashes. Pip accepted supported platform tags and rejected unsupported platforms and Python 3.9. These CI tests exercise candidate wheels and do not certify live service availability or real OS credential-store behavior.

To try local candidates without PyPI, download the successful workflow's `wheel-candidates` artifact and extract it into `dist/wheels` in your checkout. Install with `python -m pip install --no-index --find-links dist/wheels flashdata-cli==1.2.0`, or use `uv pip install` with the same options in a virtual environment. CI artifacts are retained for seven days; rebuild from the pinned release archives if they expire.

## Before publication

Require a successful native CI matrix, review contents/hashes, confirm official PyPI account/project ownership and configure a separately authorized publishing flow. Keep existing GitHub release tags/assets unchanged. Record the packaging revision separately from the native source commit. Update both languages' candidate notices and registry commands only after publication.

The preparation workflow has read-only repository permissions and no registry upload step or publishing credentials. For a native version update, release/test it first, then update the project version, pinned archives/hashes, actual platform minima and docs. Packaging-only fixes after PyPI publication require a new PEP 440 version such as `1.0.0.post1`; never overwrite uploaded wheel bytes. The current build deliberately requires package/native versions to match; a post release must explicitly adapt that guard and retain native provenance.

Support: [support@flashdata.dev](mailto:support@flashdata.dev).
