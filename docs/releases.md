# Source builds and later releases

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

[The official source repository](https://github.com/flashdatadev/flashdata-cli) is public. Version 1.0.0 is an unreleased candidate: no version tag or GitHub Release has been created. Build from a normal clone with `go mod download` and `go build ./cmd/flashdata`. No sibling SDK checkout is required.

The SDK dependency is pinned to an immutable public commit through a Go pseudo-version in go.mod. Go validates the recorded go.sum checksums. CI checks out only this repository and tests the resolved dependency; it does not silently use the sibling SDK's current working tree. Local SDK changes can be tested intentionally with a go.work workspace, outside a release build.

Run `python scripts/build_release.py --go /absolute/path/to/go` to build five CGo-disabled executables and archives with reviewed docs, examples, license notices and metadata. The script resolves the exact SDK module into Go's module cache and rejects a local module replacement. Reports retain CLI/SDK runtime source hashes, the SDK version/checksum and the Go toolchain. No upload, tag or release is created by this script.

`release-files.json` selects runtime package files. ZIP and tar metadata are normalized, and SHA256SUMS records archive checksums. Package-policy tests run with `python -m unittest discover -s scripts -p "test_*.py"`. Before distributing binaries, inspect the contents and verify the checksums. Place the matching executable in a user-controlled PATH directory and check --version/--help offline. Installed binaries require no separate runtime.

[GitHub Actions](https://github.com/flashdatadev/flashdata-cli/actions) covers Linux, Windows and native Intel macOS with Go 1.25/1.27, Apple Silicon cross-builds on both Go versions, Linux race checks, package-policy regressions and five-platform archive builds. Native Apple Silicon execution, real Secret Service/Keychain behavior, production service availability, signing and macOS notarization remain separate validation tasks. Intel macOS runners are pinned to `macos-15-intel`; the Apple Silicon hosted-runner jobs remained unassigned before test execution. Superseded CI runs are cancelled automatically. A future stable SDK tag can replace the commit pin after review; do not create SDK/CLI tags or GitHub Releases without explicit release authorization.
