# Contributing

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

## Build from source

With Go 1.25+ installed, clone the repository and build from its root. The Go SDK is resolved automatically from the version pinned in `go.mod`.

```bash
git clone https://github.com/flashdatadev/flashdata-cli.git
cd flashdata-cli
go build -trimpath -o dist/flashdata ./cmd/flashdata
./dist/flashdata --help
```

On Windows PowerShell, after cloning and entering the directory:

```powershell
go build -trimpath -o dist/flashdata.exe ./cmd/flashdata
.\dist\flashdata.exe --help
```

To build the published version instead of the latest source, add `--branch v1.0.0 --depth 1` to `git clone`. For normal installation, use the [installation guide](docs/installation.md).

## Development checks

Use Go 1.25+ and run `go mod download` to resolve the SDK revision pinned in `go.mod`; no separate SDK checkout is required. Run `go test ./...`, `go vet ./...`, `gofmt` on changed source, and race tests on a supported host. To develop both repositories together, use an explicit local `go.work` file; archive builds disable workspaces and reject SDK replacements. Do not introduce paid/live tests into the default suite. Preserve proof-before-registration, intent-before-submission, no automatic paid retries, account-bound caches, OS-lock crash safety and fresh-process verification before ACK.

All HTTP/API contract logic belongs in flashdata-go. CLI compatibility normalization, secure storage, profiles, journaling, output and host scheduling boundaries belong here. Update English/Chinese docs and runnable examples together. Publication/deployment is a separate explicit request.

With a local backend checkout available, optionally run `node scripts/check_backend_contracts.cjs /absolute/path/to/backend`. This reads the actual source schemas and invokes pure option/registration validators; it starts no server and makes no paid, registration, database or customer-cloud call. Node is used only for this development check, not by the executable or default Go tests.

After changing commands, build the local executable and run `python scripts/generate_commands.py /absolute/path/to/flashdata` to refresh the command reference. Add `--check` to verify that it matches without writing. These are optional development tools; users run only the standalone executable.

Python wheel preparation uses `packaging/pypi` and the [English packaging guide](docs/python-packaging.md) / [简体中文](docs/python-packaging.zh-CN.md). Run the packaging unit tests and native pip/uv installation checks after wrapper changes. Never introduce an installation-time binary download, shell argument forwarding or an implicit registry upload.

Agent setup embeds a reviewed snapshot from `flashdata-skills`. After committing changes in that repository, run `python scripts/sync_skills.py --source /absolute/path/to/flashdata-skills`, then `python scripts/sync_skills.py --check`. The manifest records the commit and every file hash; CI checks it offline. Do not release development snapshots marked `working-tree`. Run skill validation and tests in the source repository before refreshing the bundle.
