"""Offline regression checks for release contents and archive reproducibility."""

import io
import json
import tempfile
import unittest
import zipfile
from pathlib import Path
from unittest import mock

from release_manifest import load_release_files, write_zip_member


class ReleaseManifestTests(unittest.TestCase):
    def test_stray_private_files_are_never_selected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "examples").mkdir()
            (root / "README.md").write_text("public", encoding="utf-8")
            (root / "examples/private.json").write_text("private", encoding="utf-8")
            (root / "release-files.json").write_text('["README.md"]', encoding="utf-8")
            self.assertEqual(load_release_files(root), ["README.md"])

    def test_unsafe_paths_and_missing_files_fail_closed(self):
        cases = [
            ["../private.json"],
            ["/private.json"],
            ["C:/private.json"],
            ["examples\\private.json"],
            [".env"],
            ["storage.json"],
            ["state.job.json"],
            ["dist/output.json"],
            ["README.md", "readme.md"],
            ["missing.go"],
            [],
            {"files": []},
        ]
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "README.md").write_text("public", encoding="utf-8")
            for entries in cases:
                with self.subTest(entries=entries):
                    (root / "release-files.json").write_text(
                        json.dumps(entries), encoding="utf-8"
                    )
                    with self.assertRaises(ValueError):
                        load_release_files(root)

    def test_symlink_components_are_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "release-files.json").write_text(
                '["examples/query.json"]', encoding="utf-8"
            )
            with (
                mock.patch.object(
                    Path, "is_symlink", lambda path: path.name == "examples"
                ),
                self.assertRaises(ValueError),
            ):
                load_release_files(root)

    def test_zip_bytes_are_deterministic_and_have_safe_permissions(self):
        def build():
            output = io.BytesIO()
            with zipfile.ZipFile(output, "w") as archive:
                write_zip_member(
                    archive, "flashdata.exe", b"offline binary", executable=True
                )
                write_zip_member(archive, "README.md", b"public documentation")
            return output.getvalue()

        first, second = build(), build()
        self.assertEqual(first, second)
        with zipfile.ZipFile(io.BytesIO(first)) as archive:
            self.assertEqual(
                archive.getinfo("flashdata.exe").external_attr >> 16, 0o100755
            )
            self.assertEqual(archive.getinfo("README.md").external_attr >> 16, 0o100644)


if __name__ == "__main__":
    unittest.main()
