"""Verify/extract the native CI release candidate; never call production APIs."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tarfile
import zipfile

from release_manifest import load_release_files


def main():
    root = Path(__file__).resolve().parents[1]
    releases = root / 'dist/releases'
    name = os.environ['RELEASE_ARCHIVE']
    assert Path(name).name == name
    sums = dict(line.split('  ', 1)[::-1] for line in (releases / 'SHA256SUMS').read_text().splitlines())
    archive = releases / name
    assert hashlib.sha256(archive.read_bytes()).hexdigest() == sums[name]
    executable = 'flashdata.exe' if name.endswith('.zip') else 'flashdata'
    expected = set(load_release_files(root)) | {executable}
    destination = root / 'dist/acceptance'
    destination.mkdir(exist_ok=True)
    if name.endswith('.zip'):
        with zipfile.ZipFile(archive) as package:
            assert set(package.namelist()) == expected and len(package.namelist()) == len(expected)
            contents = {p: package.read(p) for p in expected}
    else:
        with tarfile.open(archive) as package:
            assert set(package.getnames()) == expected and len(package.getmembers()) == len(expected)
            assert all(member.isfile() for member in package.getmembers())
            contents = {p: package.extractfile(p).read() for p in expected}
    for path in expected - {executable}:
        assert contents[path] == (root / path).read_bytes(), path
    binary = destination / executable
    binary.write_bytes(contents[executable])
    binary.chmod(0o755)
    env = os.environ.copy()
    env.pop('FLASHDATA_API_KEY', None)
    report = json.loads((releases / 'build-report.json').read_text())
    actual = json.loads(subprocess.check_output([str(binary), '--version', '--json'], env=env, text=True))
    assert actual['ok'] and actual['data'] == {'version': report['version'], 'commit': report['commit']}
    for command in ['plans', 'status']:
        subprocess.run([str(binary), 'discover', command, '--help'], env=env, check=True)
    print(f'PASS: {name}; checksums, allowlist, exact docs, embedded version/commit and discovery commands')


if __name__ == '__main__':
    main()
