"""Independently validate wheel contents, metadata, RECORD and native provenance."""
import base64
import csv
import email.parser
import hashlib
import io
import json
import sys
import zipfile
from pathlib import Path

PROJECT = Path(__file__).resolve().parents[1] / "packaging/pypi"
sys.path.insert(0, str(PROJECT))
from wheel_support import manifest, safe_name, sha, verify_binary


def validate_wheel(path):
    ledger = manifest()
    with zipfile.ZipFile(path) as archive:
        names = archive.namelist()
        if len(names) != len(set(n.casefold() for n in names)):
            raise ValueError("Duplicate wheel members")
        for name in names:
            safe_name(name)
            if name.endswith((".pyc", ".pyo")) or "/__pycache__/" in name:
                raise ValueError("Compiled Python caches cannot be shipped")
        info = f"flashdata_cli-{ledger['version']}.dist-info/"
        meta = email.parser.Parser().parsestr(archive.read(info + "METADATA").decode())
        wheel = email.parser.Parser().parsestr(archive.read(info + "WHEEL").decode())
        if meta["Name"] != "flashdata-cli" or meta["Version"] != ledger["version"] or meta["Requires-Python"] != ">=3.10" or meta.get_all("Requires-Dist"):
            raise ValueError("Unexpected Python package metadata/dependencies")
        provenance = json.loads(archive.read("flashdata_cli/provenance.json"))
        target = provenance["target"]
        record = ledger["targets"][target]
        expected_name = f"flashdata_cli-{ledger['version']}-py3-none-{record['wheel_platform']}.whl"
        if Path(path).name != expected_name or wheel["Root-Is-Purelib"] != "false":
            raise ValueError("Wrong wheel filename/platform metadata")
        expected_tags = {"py3-none-" + t for t in record["wheel_platform"].split(".")}
        if set(wheel.get_all("Tag")) != expected_tags:
            raise ValueError("Wheel tags differ from the pinned platform")
        if provenance != {"version": ledger["version"], "source_commit": ledger["source_commit"], "target": target, **record}:
            raise ValueError("Wheel provenance differs from the release ledger")
        binary_name = "flashdata.exe" if target.startswith("windows/") else "flashdata"
        binary_path = "flashdata_cli/bin/" + binary_name
        binary = archive.read(binary_path)
        verify_binary(binary, target, record)
        if archive.getinfo(binary_path).external_attr >> 16 & 0o777 != 0o755:
            raise ValueError("Executable permissions lost")
        entry_points = archive.read(info + "entry_points.txt").decode()
        if "flashdata = flashdata_cli:main" not in entry_points:
            raise ValueError("Missing flashdata console entry point")
        for name in names:
            if not name.startswith(("flashdata_cli/", info)):
                raise ValueError("Unexpected top-level wheel content")
        if archive.read(info + "licenses/LICENSE") != (PROJECT / "LICENSE").read_bytes():
            raise ValueError("Missing or changed package license")
        files = set(json.loads(archive.read("flashdata_cli/share/release-files.json")))
        expected_share = {"flashdata_cli/share/" + name for name in files}
        if {n for n in names if n.startswith("flashdata_cli/share/")} != expected_share:
            raise ValueError("Release documentation or license contents differ")
        allowed = expected_share | {binary_path, "flashdata_cli/provenance.json", "flashdata_cli/__init__.py", "flashdata_cli/__main__.py"}
        allowed |= {info + name for name in ("METADATA", "WHEEL", "RECORD", "entry_points.txt", "licenses/LICENSE")}
        if set(names) != allowed:
            raise ValueError("Wheel contains unexpected files outside its public allowlist")
        rows = list(csv.reader(io.StringIO(archive.read(info + "RECORD").decode())))
        if len(rows) != len(names) or {r[0] for r in rows} != set(names):
            raise ValueError("Incomplete RECORD")
        for name, digest, size in rows:
            if name == info + "RECORD":
                if digest or size:
                    raise ValueError("RECORD must not hash itself")
                continue
            data = archive.read(name)
            expected = "sha256=" + base64.urlsafe_b64encode(hashlib.sha256(data).digest()).rstrip(b"=").decode()
            if digest != expected or int(size) != len(data):
                raise ValueError("RECORD integrity failure")
    return {"wheel": Path(path).name, "sha256": sha(Path(path).read_bytes()), "target": target, "executableSHA256": sha(binary), "recordVerified": True, "platformVerified": True}
