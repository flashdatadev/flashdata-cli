# Contributing

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Use Go 1.25+ and run `go mod download` to resolve the SDK revision pinned in `go.mod`; no separate SDK checkout is required. Run `go test ./...`, `go vet ./...`, `gofmt` on changed source, and race tests on a supported host. To develop both repositories together, use an explicit local `go.work` file; archive builds disable workspaces and reject SDK replacements. Do not introduce paid/live tests into the default suite. Preserve proof-before-registration, intent-before-submission, no automatic paid retries, account-bound caches, OS-lock crash safety and fresh-process verification before ACK.

All HTTP/API contract logic belongs in flashdata-go. CLI compatibility normalization, secure storage, profiles, journaling, output and host scheduling boundaries belong here. Update English/Chinese docs and runnable examples together. Publication/deployment is a separate explicit request.

With a local backend checkout available, optionally run `node scripts/check_backend_contracts.cjs /absolute/path/to/backend`. This reads the actual source schemas and invokes pure option/registration validators; it starts no server and makes no paid, registration, database or customer-cloud call. Node is used only for this development check, not by the executable or default Go tests.

After changing commands, build the local executable and run `python scripts/generate_commands.py /absolute/path/to/flashdata` to refresh the command reference. Add `--check` to verify that it matches without writing. These are optional development tools; users run only the standalone executable.

Python wheel preparation uses `packaging/pypi` and the [English packaging guide](docs/python-packaging.md) / [简体中文](docs/python-packaging.zh-CN.md). Run the packaging unit tests and native pip/uv installation checks after wrapper changes. Never introduce an installation-time binary download, shell argument forwarding or an implicit registry upload.
