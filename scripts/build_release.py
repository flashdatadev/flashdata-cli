"""Build local standalone archives and checksums; never publish or upload."""

import argparse
import gzip
import hashlib
import io
import json
import os
import re
import subprocess
import tarfile
import zipfile
from pathlib import Path

from release_manifest import load_release_files, write_zip_member

ROOT = Path(__file__).resolve().parents[1]
TARGETS = [
    ("linux", "amd64"),
    ("linux", "arm64"),
    ("darwin", "amd64"),
    ("darwin", "arm64"),
    ("windows", "amd64"),
]


def source_digest(root):
    files = [root / "go.mod", root / "go.sum"]
    for directory in [root / "cmd", root / "internal"]:
        files.extend(
            p for p in directory.rglob("*.go") if not p.name.endswith("_test.go")
        )
    files.extend(p for p in root.glob("*.go") if not p.name.endswith("_test.go"))
    schemas = root / "internal/contract/schemas.json"
    if schemas.is_file():
        files.append(schemas)
    digest = hashlib.sha256()
    for path in sorted(files):
        digest.update(
            path.relative_to(root).as_posix().encode()
            + b"\0"
            + path.read_bytes()
            + b"\0"
        )
    return digest.hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--go", default="go")
    parser.add_argument("--version", default="1.1.2")
    parser.add_argument("--commit", default="local", help="full source Git commit")
    args = parser.parse_args()
    if not re.fullmatch(r"\d+\.\d+\.\d+(?:-[A-Za-z0-9.-]+)?", args.version):
        parser.error("version must be a semantic version")
    if not re.fullmatch(r"local|[0-9a-fA-F]{40}", args.commit):
        parser.error("commit must be 'local' or a full 40-character Git commit")
    args.commit = args.commit.lower()
    notes = ROOT / "docs" / f"release-notes-{args.version}.md"
    if not notes.is_file():
        parser.error(f"missing versioned release notes: {notes.name}")
    go = args.go
    if Path(go).exists():
        go = str(Path(go).resolve())
    binaries = ROOT / "dist/binaries"
    releases = ROOT / "dist/releases"
    binaries.mkdir(parents=True, exist_ok=True)
    releases.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env.setdefault("GOCACHE", str(ROOT / ".cache/go-build"))
    env["CGO_ENABLED"] = "0"
    env["GOWORK"] = "off"
    sdk_module = "github.com/flashdatadev/flashdata-go"
    resolved = json.loads(
        subprocess.check_output(
            [go, "list", "-mod=readonly", "-m", "-json", sdk_module],
            cwd=ROOT,
            env=env,
            text=True,
        )
    )
    if resolved.get("Replace"):
        parser.error(
            "release builds require the pinned SDK module without a local replacement"
        )
    sdk = json.loads(
        subprocess.check_output(
            [go, "mod", "download", "-json", sdk_module],
            cwd=ROOT,
            env=env,
            text=True,
        )
    )
    sdk_root = Path(sdk["Dir"])
    shared = [(ROOT / name, name) for name in load_release_files(ROOT)]
    source_hashes = {
        "cli": source_digest(ROOT),
        "sdk": source_digest(sdk_root),
    }
    toolchain = subprocess.check_output([go, "version"], env=env, text=True).strip()
    artifacts = []
    for system, arch in TARGETS:
        env.update({"GOOS": system, "GOARCH": arch})
        executable = "flashdata.exe" if system == "windows" else "flashdata"
        binary = (
            binaries
            / f"flashdata-{system}-{arch}{'.exe' if system == 'windows' else ''}"
        )
        subprocess.run(
            [
                go,
                "build",
                "-trimpath",
                "-buildvcs=false",
                "-mod=readonly",
                "-ldflags",
                f"-s -w -X main.version={args.version} -X main.commit={args.commit}",
                "-o",
                str(binary),
                "./cmd/flashdata",
            ],
            cwd=ROOT,
            env=env,
            check=True,
        )
        files = [(binary, executable), *shared]
        archive = (
            releases
            / f"flashdata_{args.version}_{system}_{arch}{'.zip' if system == 'windows' else '.tar.gz'}"
        )
        if system == "windows":
            with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED) as z:
                for path, name in files:
                    write_zip_member(
                        z, name, path.read_bytes(), executable=name == executable
                    )
        else:
            with (
                archive.open("wb") as output,
                gzip.GzipFile(
                    filename="", fileobj=output, mode="wb", mtime=0
                ) as compressed,
                tarfile.open(fileobj=compressed, mode="w") as tar,
            ):
                for path, name in files:
                    data = path.read_bytes()
                    info = tarfile.TarInfo(name)
                    info.size = len(data)
                    info.mode = 0o755 if name == executable else 0o644
                    info.mtime = 0
                    tar.addfile(info, io.BytesIO(data))
        artifacts.append(
            {
                "platform": f"{system}/{arch}",
                "archive": archive.name,
                "sha256": hashlib.sha256(archive.read_bytes()).hexdigest(),
            }
        )
        print(f"Built {system}/{arch}: {archive.name}", flush=True)
    (releases / "RELEASE_NOTES.md").write_bytes(notes.read_bytes())
    (releases / "build-report.json").write_text(
        json.dumps(
            {
                "version": args.version,
                "commit": args.commit,
                "publishes": False,
                "cgoEnabled": False,
                "allowlisted": True,
                "sourceSHA256": source_hashes,
                "sdkModule": {
                    "path": sdk_module,
                    "version": sdk["Version"],
                    "sum": sdk["Sum"],
                },
                "goToolchain": toolchain,
                "artifacts": artifacts,
            },
            indent=2,
        )
        + "\n",
        encoding="utf-8",
        newline="\n",
    )
    checksums = [(a["archive"], a["sha256"]) for a in artifacts]
    for name in ["build-report.json", "RELEASE_NOTES.md"]:
        checksums.append(
            (name, hashlib.sha256((releases / name).read_bytes()).hexdigest())
        )
    (releases / "SHA256SUMS").write_text(
        "".join(f"{digest}  {name}\n" for name, digest in sorted(checksums)),
        encoding="utf-8",
        newline="\n",
    )


if __name__ == "__main__":
    main()
