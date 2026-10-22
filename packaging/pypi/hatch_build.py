"""Hatch hook: put a pinned native CLI into an explicitly tagged wheel."""
import json
import io
import os
import sys
import tempfile
import zipfile
from pathlib import Path

from hatchling.builders.hooks.plugin.interface import BuildHookInterface

# Hatch loads custom hooks by filename; explicitly resolve this sibling module.
sys.path.insert(0, str(Path(__file__).resolve().parent))
from wheel_support import host_target, verified_files


class NativeCLIHook(BuildHookInterface):
    def initialize(self, version, build_data):
        if version != "standard":
            raise ValueError("Editable CLI wheels are unsupported; build a standard wheel")
        target = os.environ.get("FLASHDATA_WHEEL_TARGET") or host_target()
        archives = os.environ.get("FLASHDATA_CLI_ARCHIVE_DIR", str(Path(self.root) / ".cache/archives"))
        ledger, record, files = verified_files(target, archives, os.environ.get("FLASHDATA_WHEEL_DOWNLOAD") == "1")
        if self.metadata.version != ledger["version"]:
            raise ValueError("Python package and pinned CLI versions differ")
        self._stage = tempfile.TemporaryDirectory(prefix="flashdata-wheel-")
        stage = Path(self._stage.name)
        binary_name = "flashdata.exe" if target.startswith("windows/") else "flashdata"
        for name, data in files.items():
            relative = "bin/" + name if name == binary_name else "share/" + name
            destination = stage / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes(data)
            destination.chmod(0o755 if name == binary_name else 0o644)
            build_data["force_include"][str(destination)] = "flashdata_cli/" + relative
        provenance = stage / "provenance.json"
        provenance.write_text(json.dumps({"version": ledger["version"], "source_commit": ledger["source_commit"], "target": target, **record}, indent=2) + "\n", encoding="utf-8", newline="\n")
        build_data["force_include"][str(provenance)] = "flashdata_cli/provenance.json"
        build_data["tag"] = "py3-none-" + record["wheel_platform"]
        build_data["pure_python"] = False

    def finalize(self, version, build_data, artifact_path):
        try:
            # Windows chmod cannot preserve Unix executable bits. Encode them
            # explicitly so cross-built Linux/macOS wheels remain runnable.
            path = Path(artifact_path)
            output = io.BytesIO()
            with zipfile.ZipFile(path) as source, zipfile.ZipFile(output, "w") as destination:
                for entry in source.infolist():
                    entry.create_system = 3
                    executable = entry.filename.startswith("flashdata_cli/bin/")
                    entry.external_attr = (0o100755 if executable else 0o100644) << 16
                    destination.writestr(entry, source.read(entry), compresslevel=9)
            path.write_bytes(output.getvalue())
        finally:
            self._stage.cleanup()
