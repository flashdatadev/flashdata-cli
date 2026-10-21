"""Read a reviewed file allowlist and write deterministic ZIP members."""

import json
from pathlib import Path, PurePosixPath
from zipfile import ZIP_DEFLATED, ZipInfo


def load_release_files(root):
    root = Path(root).resolve()
    entries = json.loads((root / "release-files.json").read_text(encoding="utf-8"))
    if not isinstance(entries, list) or not entries:
        raise ValueError("release-files.json must contain a nonempty file list")
    seen = set()
    for name in entries:
        if not isinstance(name, str) or not name or "\\" in name:
            raise ValueError("invalid release path")
        path = PurePosixPath(name)
        if (
            path.is_absolute()
            or path.as_posix() != name
            or any(part in {".", ".."} or ":" in part for part in path.parts)
            or any(
                part
                in {".git", ".cache", ".tools", "dist", "__pycache__", "node_modules"}
                or part.startswith(".env")
                for part in path.parts
            )
            or path.name == "storage.json"
            or name.endswith(".job.json")
            or name.casefold() in seen
        ):
            raise ValueError("unsafe or duplicate release path")
        seen.add(name.casefold())
        candidate = root
        for part in path.parts:
            candidate = candidate / part
            if candidate.is_symlink():
                raise ValueError("release files cannot be symbolic links")
        if not candidate.is_file() or not candidate.resolve().is_relative_to(root):
            raise ValueError("release file is missing or outside the checkout")
    return sorted(entries)


def write_zip_member(archive, name, data, executable=False):
    info = ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
    info.create_system = 3
    info.external_attr = (0o100755 if executable else 0o100644) << 16
    info.compress_type = ZIP_DEFLATED
    archive.writestr(info, data, compresslevel=9)
