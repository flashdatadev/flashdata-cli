"""Bundle a reviewed skills checkout; offline --check validates the pinned snapshot."""
import argparse
import hashlib
import json
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BUNDLE = ROOT / 'internal/agentskills/bundle'


def snapshot(source):
    result = {}
    for file in sorted((source / 'skills').rglob('*')):
        if file.is_symlink():
            raise ValueError(f'Symlink in skill source: {file}')
        if not file.is_file() or '__pycache__' in file.parts:
            continue
        if file.suffix not in ('.md', '.py', '.json', '.yaml'):
            raise ValueError(f'Unexpected skill file: {file}')
        result[file.relative_to(source).as_posix()] = file.read_bytes().replace(b'\r\n', b'\n')
    if len([name for name in result if name.endswith('/SKILL.md')]) != 17:
        raise ValueError('Expected 17 complete skills')
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--source', type=Path)
    parser.add_argument('--check', action='store_true')
    parser.add_argument('--allow-dirty', action='store_true', help='development only; never release a dirty snapshot')
    args = parser.parse_args()
    if args.check:
        manifest = json.loads((BUNDLE / 'source.json').read_text(encoding='utf-8'))
        if not re.fullmatch(r'[a-f0-9]{40}', manifest['commit']):
            raise ValueError('Bundle must pin a committed skills revision')
        actual = snapshot(BUNDLE)
        hashes = {name: hashlib.sha256(data).hexdigest() for name, data in actual.items()}
        if hashes != manifest['files']:
            raise ValueError('Bundled skills differ from their pinned manifest')
        if args.source and actual != snapshot(args.source.resolve()):
            raise ValueError('Bundle differs from the supplied skills checkout')
        print(f'Verified {len(actual)} bundled files from {manifest["commit"]}')
        return
    if not args.source:
        parser.error('--source is required when updating')
    source = args.source.resolve()
    status = subprocess.check_output(['git', 'status', '--porcelain', '--', 'skills'], cwd=source, text=True)
    if status and not args.allow_dirty:
        raise ValueError('Commit and review skill changes before bundling')
    commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=source, text=True).strip()
    files = snapshot(source)
    if status:
        commit = 'working-tree'
    expected = set(files) | {'source.json'}
    # Remove only files inside the verified bundle root, never a computed checkout.
    for path in BUNDLE.rglob('*'):
        if path.is_symlink():
            raise ValueError(f'Unexpected bundle symlink: {path}')
        if path.is_file() and path.relative_to(BUNDLE).as_posix() not in expected:
            path.unlink()
    for name, data in files.items():
        path = BUNDLE / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
    manifest = {'repository':'https://github.com/flashdatadev/flashdata-skills', 'commit':commit,
                'files':{name:hashlib.sha256(data).hexdigest() for name,data in files.items()}}
    (BUNDLE/'source.json').write_text(json.dumps(manifest,indent=2)+'\n',encoding='utf-8')
    print(f'Bundled {len(files)} files from {commit}')


if __name__ == '__main__':
    main()
