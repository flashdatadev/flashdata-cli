# FlashData CLI v1.3.0

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Transcript and captions now support several languages in one request:

```sh
flashdata youtube transcript dQw4w9WgXcQ --language en,es
flashdata youtube captions dQw4w9WgXcQ --language en --language es
```

Omit `--language` to prefer `en-orig`, then `en`, returning one track. One explicit
code keeps the single-track response. Several codes return available tracks and
per-language errors. JSON requests also support a one-element language array.

The release pins Go SDK v1.2.0 and refreshes the bundled subtitle skills and
examples. Run `flashdata setup` after upgrading to refresh unmodified skills.
Bulk exports remain JSONL, and other API contracts are unchanged.

Standalone archives support Linux amd64/arm64, macOS Intel/Apple Silicon and Windows
x64. Homebrew and the signed APT/RPM repositories distribute the same native builds.
Python installer wheels are supplied separately for manual PyPI upload. Until the
new version appears on PyPI, use the [native installation options](installation.md).
