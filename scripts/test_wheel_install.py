"""Offline native installation checks for pip, uv pip and uv tool; never publish."""
import argparse
import json
import os
import subprocess
import sys
from pathlib import Path

from wheel_validation import manifest, validate_wheel
from wheel_support import host_target, sha

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--wheels", type=Path, default=ROOT / "dist/wheels")
    parser.add_argument("--uv", required=True, help="Path to uv executable")
    parser.add_argument("--python", default=sys.executable)
    parser.add_argument("--work-dir", type=Path, default=ROOT / ".cache/wheel-install-tests")
    args = parser.parse_args()
    work = args.work_dir.resolve()
    if work.exists() and any(work.iterdir()):
        parser.error("Use an empty test work directory")
    work.mkdir(parents=True, exist_ok=True)
    wheels = args.wheels.resolve()
    uv = str(Path(args.uv).resolve()) if Path(args.uv).exists() else args.uv
    ledger, target = manifest(), host_target()
    matching = [p for p in wheels.glob("*.whl") if validate_wheel(p)["target"] == target]
    if len(matching) != 1:
        raise ValueError("Expected exactly one matching verified wheel")
    env = {**os.environ, "UV_OFFLINE": "1", "UV_NO_CONFIG": "1", "UV_CACHE_DIR": str(work / "uv-cache"),
           "UV_TOOL_DIR": str(work / "uv-tools"), "UV_TOOL_BIN_DIR": str(work / "uv-bin"),
           "PIP_CONFIG_FILE": os.devnull, "PIP_NO_INDEX": "1", "PIP_DISABLE_PIP_VERSION_CHECK": "1", "PYTHONIOENCODING": "utf-8"}
    for variable in ("FLASHDATA_API_KEY", "PYTHONPATH", "PYTHONHOME", "VIRTUAL_ENV", "PIP_TARGET", "PIP_PREFIX", "PIP_USER", "UV_PYTHON", "UV_PROJECT_ENVIRONMENT"):
        env.pop(variable, None)

    def run(command, *, check=True, data=None):
        result = subprocess.run([str(c) for c in command], input=data, capture_output=True, encoding="utf-8", env=env, timeout=180)
        if check and result.returncode:
            raise RuntimeError(result.stdout + result.stderr)
        return result

    python = Path(run([args.python, "-c", "import sys;print(sys.executable)"]).stdout.strip()).resolve()
    uv_version = run([uv, "--version"]).stdout.strip()
    reports = []
    for method in ("pip", "uv-pip", "uv-tool"):
        folder = work / method
        marker = folder / "user-state/do-not-remove.txt"
        marker.parent.mkdir(parents=True)
        marker.write_text("preserved profile/run state\n", encoding="utf-8")
        if method == "uv-tool":
            command = [uv, "tool", "install", "--python", python, "--no-index", "--find-links", wheels, "flashdata-cli==1.0.0"]
            run(command)
            executable = work / "uv-bin" / ("flashdata.exe" if os.name == "nt" else "flashdata")
            environment = work / "uv-tools/flashdata-cli"
        else:
            environment = folder / "venv"
            run([python, "-m", "venv", environment])
            interpreter = environment / ("Scripts/python.exe" if os.name == "nt" else "bin/python")
            if method == "pip":
                command = [interpreter, "-m", "pip", "install", "--no-index", "--find-links", wheels, "flashdata-cli==1.0.0"]
            else:
                command = [uv, "pip", "install", "--python", interpreter, "--no-index", "--find-links", wheels, "flashdata-cli==1.0.0"]
            run(command)
            executable = environment / ("Scripts/flashdata.exe" if os.name == "nt" else "bin/flashdata")
        interpreter = environment / ("Scripts/python.exe" if os.name == "nt" else "bin/python")
        package = Path(run([interpreter, "-c", "import pathlib,flashdata_cli;print(pathlib.Path(flashdata_cli.__file__).parent)"]).stdout.strip())
        binary = package / "bin" / ("flashdata.exe" if os.name == "nt" else "flashdata")
        if sha(binary.read_bytes()) != ledger["targets"][target]["executable_sha256"]:
            raise ValueError("Installer modified the upstream executable")
        version = json.loads(run([executable, "--version", "--json"]).stdout)
        assert version["ok"] and version["data"]["version"] == ledger["version"] and version["data"]["commit"] == ledger["source_commit"]
        assert "bulk" in run([executable, "--help"]).stdout
        commands = json.loads(run([executable, "commands", "--json"]).stdout)
        assert {"auth", "bulk", "query", "runs"} <= {c["name"] for c in commands["data"]["commands"]}
        assert run([interpreter, "-m", "flashdata_cli", "--version", "--json"]).stdout == run([binary, "--version", "--json"]).stdout
        for native_args, stdin in [(["unknown-command-中文;literal", "--json"], None), (["query", "--input", "-", "--non-interactive", "--state-dir", str(folder / "isolated-state")], "{invalid-json\n")]:
            native = run([binary, *native_args], check=False, data=stdin)
            wrapper = run([executable, *native_args], check=False, data=stdin)
            assert native.returncode != 0 and wrapper.returncode == native.returncode
            assert wrapper.stdout == native.stdout and wrapper.stderr == native.stderr
        assert (package / "share/README.zh-CN.md").is_file()
        assert (package / "share/examples/query.json").is_file()
        assert len(list((package / "share/licenses").glob("*.txt"))) >= 10
        # Reinstall the same candidate using the same manager; no global PATH edits.
        reinstall = command + (["--reinstall"] if method != "pip" else ["--force-reinstall"])
        run(reinstall)
        assert json.loads(run([executable, "--version", "--json"]).stdout)["data"]["version"] == ledger["version"]
        if method == "uv-tool":
            run([uv, "tool", "uninstall", "flashdata-cli"])
        elif method == "pip":
            run([interpreter, "-m", "pip", "uninstall", "-y", "flashdata-cli"])
        else:
            run([uv, "pip", "uninstall", "--python", interpreter, "flashdata-cli"])
        assert not executable.exists() and marker.read_text(encoding="utf-8") == "preserved profile/run state\n"
        reports.append({"method": method, "install": True, "reinstall": True, "uninstall": True, "offlineVersionHelpCommands": True, "streamsArgumentsExitCode": True, "exactNativeBinary": True, "userStatePreserved": True})
        print("PASS: " + method, flush=True)
    report = {"target": target, "python": run([python, "--version"]).stdout.strip(), "uv": uv_version, "publishes": False, "methods": reports}
    (work / "report.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8", newline="\n")
    print(json.dumps(report))


if __name__ == "__main__":
    main()
