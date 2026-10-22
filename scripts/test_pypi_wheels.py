"""Offline unit regressions for binary packaging boundaries and CLI forwarding."""
import importlib.util
import io
import json
import os
import struct
import sys
import tarfile
import tempfile
import unittest
import zipfile
from pathlib import Path
from unittest import mock

from wheel_validation import PROJECT, validate_wheel
from wheel_support import manifest, read_archive, sha, verified_files, verify_binary

spec = importlib.util.spec_from_file_location("wheel_launcher", PROJECT / "src/flashdata_cli/__init__.py")
launcher = importlib.util.module_from_spec(spec)
spec.loader.exec_module(launcher)


class WheelTests(unittest.TestCase):
    def test_project_license_and_translated_command_blocks(self):
        import re
        root = PROJECT.parents[1]
        self.assertEqual((PROJECT / "LICENSE").read_bytes(), (root / "LICENSE").read_bytes())
        for english, chinese in [(PROJECT / "README.md", PROJECT / "README.zh-CN.md"), (root / "docs/python-packaging.md", root / "docs/python-packaging.zh-CN.md")]:
            blocks = lambda p: re.findall(r"```[^\n]*\n(.*?)```", p.read_text(encoding="utf-8"), re.S)
            self.assertEqual(blocks(english), blocks(chinese))

    def test_unexpected_private_file_in_wheel_is_rejected(self):
        candidates = list((PROJECT.parents[1] / "dist/wheels").glob("*.whl"))
        if not candidates:
            self.skipTest("Build wheel candidates before testing artifact mutation")
        with tempfile.TemporaryDirectory() as directory:
            source = candidates[0]
            destination = Path(directory) / source.name
            destination.write_bytes(source.read_bytes())
            with zipfile.ZipFile(destination, "a") as archive:
                archive.writestr("flashdata_cli/private-config.json", "not for distribution")
            with self.assertRaisesRegex(ValueError, "allowlist"):
                validate_wheel(destination)

    def test_archive_paths_links_and_duplicate_names_rejected(self):
        for name, kind in [("../private", tarfile.REGTYPE), ("C:/private", tarfile.REGTYPE), ("safe", tarfile.SYMTYPE), ("safe", tarfile.DIRTYPE)]:
            data = io.BytesIO()
            with tarfile.open(fileobj=data, mode="w:gz") as archive:
                entry = tarfile.TarInfo(name)
                entry.type = kind
                archive.addfile(entry)
            with self.subTest(name=name, kind=kind), self.assertRaises(ValueError):
                read_archive(data.getvalue(), False)
        data = io.BytesIO()
        with tarfile.open(fileobj=data, mode="w:gz") as archive:
            archive.addfile(tarfile.TarInfo("same"))
            archive.addfile(tarfile.TarInfo("SAME"))
        with self.assertRaisesRegex(ValueError, "Duplicate"):
            read_archive(data.getvalue(), False)

    def test_missing_and_changed_archive_never_download_implicitly(self):
        with tempfile.TemporaryDirectory() as directory, mock.patch("urllib.request.urlopen") as network:
            with self.assertRaisesRegex(ValueError, "archive missing"):
                verified_files("windows/amd64", directory)
            path = Path(directory) / manifest()["targets"]["windows/amd64"]["archive"]
            path.write_bytes(b"corrupt")
            with self.assertRaisesRegex(ValueError, "checksum mismatch"):
                verified_files("windows/amd64", directory, download=True)
            network.assert_not_called()

    def test_elf_architecture_and_dynamic_linking_rejected(self):
        binary = bytearray(128)
        binary[:6] = b"\x7fELF\x02\x01"
        struct.pack_into("<H", binary, 18, 183)
        with self.assertRaisesRegex(ValueError, "architecture"):
            verify_binary(binary, "linux/amd64", {"executable_sha256": sha(binary)})
        struct.pack_into("<H", binary, 18, 62)
        struct.pack_into("<Q", binary, 32, 64)
        struct.pack_into("<HH", binary, 54, 56, 1)
        struct.pack_into("<I", binary, 64, 3)
        with self.assertRaisesRegex(ValueError, "static"):
            verify_binary(binary, "linux/amd64", {"executable_sha256": sha(binary)})

    def test_windows_launcher_passes_literal_arguments_and_preserves_exit_status(self):
        child = mock.Mock()
        child.wait.side_effect = [KeyboardInterrupt(), 23]
        path = PROJECT / "src/flashdata_cli/__init__.py"
        with mock.patch.object(launcher.os, "name", "nt"), mock.patch.object(launcher, "Path", return_value=path), mock.patch.object(launcher.subprocess, "Popen", return_value=child) as process, mock.patch.object(sys, "argv", ["flashdata", "中文 path;literal", "--json"]):
            self.assertEqual(launcher.main(), 23)
            arguments = process.call_args.args[0]
            self.assertEqual(arguments[1:], ["中文 path;literal", "--json"])
            self.assertEqual(process.call_args.kwargs, {})
            self.assertTrue(arguments[0].endswith("flashdata.exe"))

    def test_launcher_startup_error_does_not_echo_arguments(self):
        with mock.patch.object(launcher.os, "name", "nt"), mock.patch.object(launcher, "Path", return_value=PROJECT / "src/flashdata_cli/__init__.py"), mock.patch.object(launcher.subprocess, "Popen", side_effect=FileNotFoundError(2, "not found")), mock.patch.object(sys, "argv", ["flashdata", "secret-argument"]), mock.patch.object(sys, "stderr", new_callable=io.StringIO) as error:
            self.assertEqual(launcher.main(), 127)
            self.assertNotIn("secret-argument", error.getvalue())

    def test_posix_exec_replaces_process_without_shell_or_stream_redirection(self):
        class Replaced(Exception):
            pass
        path = PROJECT / "src/flashdata_cli/__init__.py"
        with mock.patch.object(launcher.os, "name", "posix"), mock.patch.object(launcher, "Path", return_value=path), mock.patch.object(launcher.os, "execv", side_effect=Replaced) as execute, mock.patch.object(sys, "argv", ["flashdata", "bulk", "--input", "two words.json"]):
            with self.assertRaises(Replaced):
                launcher.main()
            binary, arguments = execute.call_args.args
            self.assertTrue(binary.endswith("bin/flashdata") or binary.endswith("bin\\flashdata"))
            self.assertEqual(arguments[1:], ["bulk", "--input", "two words.json"])


if __name__ == "__main__":
    unittest.main()
