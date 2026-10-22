"""Build verified local PyPI wheels; this script never uploads or publishes."""
import argparse
import hashlib
import importlib.metadata
import json
import os
import platform
import subprocess
import sys
from pathlib import Path

from wheel_validation import PROJECT, manifest, validate_wheel

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--target", action="append", choices=list(manifest()["targets"]))
    parser.add_argument("--archive-dir", type=Path, default=ROOT / ".cache/wheel-archives")
    parser.add_argument("--download", action="store_true", help="Explicitly download missing checksum-pinned release archives")
    parser.add_argument("--output", type=Path, default=ROOT / "dist/wheels")
    args = parser.parse_args()
    output = args.output.resolve()
    if output.exists() and any(output.iterdir()):
        parser.error("Use an empty output directory to avoid overwriting wheel candidates")
    output.mkdir(parents=True, exist_ok=True)
    reports = []
    for target in args.target or list(manifest()["targets"]):
        env = {**os.environ, "FLASHDATA_WHEEL_TARGET": target, "FLASHDATA_CLI_ARCHIVE_DIR": str(args.archive_dir.resolve()), "FLASHDATA_WHEEL_DOWNLOAD": "1" if args.download else "0"}
        subprocess.run([sys.executable, "-m", "build", "--wheel", "--no-isolation", "--outdir", str(output), str(PROJECT)], env=env, check=True)
        name = f"flashdata_cli-{manifest()['version']}-py3-none-{manifest()['targets'][target]['wheel_platform']}.whl"
        reports.append(validate_wheel(output / name))
    sources = [PROJECT / name for name in ("pyproject.toml", "README.md", "README.zh-CN.md", "LICENSE", "upstream.json", "hatch_build.py", "wheel_support.py", "requirements-build.txt")]
    sources += [PROJECT / "src/flashdata_cli" / name for name in ("__init__.py", "__main__.py")]
    sources += [Path(__file__).resolve(), Path(__file__).with_name("wheel_validation.py")]
    digest = hashlib.sha256()
    for path in sorted(sources):
        digest.update(path.relative_to(ROOT).as_posix().encode() + b"\0" + path.read_bytes() + b"\0")
    report = {"version": manifest()["version"], "nativeSourceCommit": manifest()["source_commit"], "packagingSourceSHA256": digest.hexdigest(), "publishes": False, "python": platform.python_version(), "hatchling": importlib.metadata.version("hatchling"), "artifacts": reports}
    (output / "build-report.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8", newline="\n")
    (output / "SHA256SUMS").write_text("".join(f"{r['sha256']}  {r['wheel']}\n" for r in reports), encoding="utf-8", newline="\n")
    print(json.dumps(report))


if __name__ == "__main__":
    main()
