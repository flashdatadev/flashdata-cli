"""Exercise pip's real wheel selector for supported and unsupported platforms."""
import argparse
import json
import os
import subprocess
import sys
from pathlib import Path

from wheel_validation import manifest, validate_wheel

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--wheels", type=Path, default=ROOT / "dist/wheels")
    parser.add_argument("--work-dir", type=Path, default=ROOT / ".cache/wheel-resolution")
    args = parser.parse_args()
    work = args.work_dir.resolve()
    if work.exists() and any(work.iterdir()):
        parser.error("Use an empty resolution work directory")
    work.mkdir(parents=True, exist_ok=True)
    supported = [(p, target, "3.10") for target, record in manifest()["targets"].items() for p in record["wheel_platform"].split(".")]
    rejected = [(p, None, "3.10") for p in ("musllinux_1_2_x86_64", "win_arm64", "macosx_12_0_arm64", "manylinux_2_17_i686")]
    rejected.append(("win_amd64", None, "3.9"))
    env = {**os.environ, "PIP_CONFIG_FILE": os.devnull, "PIP_DISABLE_PIP_VERSION_CHECK": "1"}
    report = []
    for number, (platform, target, python) in enumerate(supported + rejected):
        destination = work / str(number)
        command = [sys.executable, "-m", "pip", "download", "--no-index", "--find-links", str(args.wheels.resolve()), "--only-binary=:all:", "--no-deps", "--platform", platform, "--python-version", python, "--implementation", "cp", "--abi", "none", "--dest", str(destination), f"flashdata-cli=={manifest()['version']}"]
        result = subprocess.run(command, capture_output=True, text=True, encoding="utf-8", env=env, timeout=120)
        (work / (str(number) + ".log")).write_text(result.stdout + result.stderr, encoding="utf-8")
        if target:
            if result.returncode:
                raise RuntimeError(result.stdout + result.stderr)
            candidates = list(destination.glob("*.whl"))
            assert len(candidates) == 1 and validate_wheel(candidates[0])["target"] == target
        else:
            reason = "requires a different Python" if python == "3.9" else "No matching distribution found"
            assert result.returncode != 0 and reason in result.stderr, (platform, python, result.stdout, result.stderr)
        report.append({"platform": platform, "python": python, "expectedTarget": target, "passed": True})
    (work / "report.json").write_text(json.dumps({"publishes": False, "checks": report}, indent=2) + "\n", encoding="utf-8", newline="\n")
    print(json.dumps(report))


if __name__ == "__main__":
    main()
