"""Verify pinned release assets; never execute a foreign binary while packaging."""
import hashlib
import io
import json
import platform
import stat
import struct
import tarfile
import urllib.request
import zipfile
from pathlib import Path, PurePosixPath

ROOT = Path(__file__).resolve().parent


def sha(data):
    return hashlib.sha256(data).hexdigest()


def manifest():
    return json.loads((ROOT / "upstream.json").read_text(encoding="utf-8"))


def host_target():
    system = {"Linux": "linux", "Darwin": "darwin", "Windows": "windows"}.get(platform.system())
    arch = {"AMD64": "amd64", "x86_64": "amd64", "aarch64": "arm64", "arm64": "arm64"}.get(platform.machine())
    target = f"{system}/{arch}"
    if target not in manifest()["targets"]:
        raise ValueError("Unsupported FlashData wheel host; use a supported standalone release")
    return target


def safe_name(name):
    path = PurePosixPath(name)
    if not name or "\\" in name or path.is_absolute() or path.as_posix() != name or any(part in (".", "..") or ":" in part for part in path.parts):
        raise ValueError("Unsafe release archive member")
    return name


def read_archive(raw, windows):
    files = {}
    def add(name, data):
        safe_name(name)
        if name.casefold() in {n.casefold() for n in files}:
            raise ValueError("Duplicate release archive member")
        files[name] = data
    if windows:
        with zipfile.ZipFile(io.BytesIO(raw)) as archive:
            for info in archive.infolist():
                mode = info.external_attr >> 16
                if info.is_dir() or (stat.S_IFMT(mode) not in (0, stat.S_IFREG)):
                    raise ValueError("Release archive must contain only regular files")
                add(info.filename, archive.read(info))
    else:
        with tarfile.open(fileobj=io.BytesIO(raw), mode="r:gz") as archive:
            for info in archive.getmembers():
                if not info.isfile():
                    raise ValueError("Release archive must contain only regular files")
                add(info.name, archive.extractfile(info).read())
    return files


def verify_binary(binary, target, record):
    if sha(binary) != record["executable_sha256"]:
        raise ValueError("Upstream executable checksum mismatch")
    system, arch = target.split("/")
    if system == "linux":
        if binary[:6] != b"\x7fELF\x02\x01" or struct.unpack_from("<H", binary, 18)[0] != {"amd64": 62, "arm64": 183}[arch]:
            raise ValueError("Wrong ELF architecture")
        offset = struct.unpack_from("<Q", binary, 32)[0]
        size, count = struct.unpack_from("<HH", binary, 54)
        if any(struct.unpack_from("<I", binary, offset + size * n)[0] == 3 for n in range(count)):
            raise ValueError("Linux wheel requires a static executable without PT_INTERP")
    elif system == "windows":
        offset = struct.unpack_from("<I", binary, 60)[0]
        if binary[:2] != b"MZ" or binary[offset:offset+4] != b"PE\0\0" or struct.unpack_from("<H", binary, offset+4)[0] != 0x8664:
            raise ValueError("Wrong PE architecture")
    else:
        if binary[:4] != b"\xcf\xfa\xed\xfe" or struct.unpack_from("<I", binary, 4)[0] != {"amd64": 0x1000007, "arm64": 0x100000c}[arch]:
            raise ValueError("Wrong Mach-O architecture")
        offset, minimum = 32, None
        for _ in range(struct.unpack_from("<I", binary, 16)[0]):
            command, size = struct.unpack_from("<II", binary, offset)
            if command == 0x32:
                minimum = struct.unpack_from("<I", binary, offset+12)[0]
            elif command == 0x24:
                minimum = struct.unpack_from("<I", binary, offset+8)[0]
            offset += size
        # Do not advertise compatibility below the actual Mach-O deployment target.
        if minimum is None or not record["wheel_platform"].startswith(f"macosx_{minimum >> 16}_0_"):
            raise ValueError("Wheel macOS minimum differs from the executable")


def verified_files(target, archive_dir, download=False):
    ledger = manifest()
    record = ledger["targets"][target]
    path = Path(archive_dir).resolve() / record["archive"]
    if not path.exists():
        if not download:
            raise ValueError("Pinned archive missing; supply --archive-dir or explicitly enable --download")
        url = f"https://github.com/flashdatadev/flashdata-cli/releases/download/v{ledger['version']}/{record['archive']}"
        with urllib.request.urlopen(url, timeout=90) as response:
            raw = response.read()
        if sha(raw) != record["archive_sha256"]:
            raise ValueError("Downloaded archive checksum mismatch")
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(raw)
    raw = path.read_bytes()
    if sha(raw) != record["archive_sha256"]:
        raise ValueError("Upstream archive checksum mismatch")
    files = read_archive(raw, target.startswith("windows/"))
    binary_name = "flashdata.exe" if target.startswith("windows/") else "flashdata"
    verify_binary(files[binary_name], target, record)
    expected = set(json.loads(files["release-files.json"])) | {binary_name}
    if set(files) != expected:
        raise ValueError("Upstream archive differs from its reviewed file allowlist")
    return ledger, record, files
