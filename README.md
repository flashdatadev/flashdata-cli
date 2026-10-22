# FlashData CLI

[简体中文](README.zh-CN.md) · [FlashData](https://flashdata.dev) · [API docs](https://flashdata.dev/docs) · [Pricing](https://flashdata.dev/pricing) · [Release v1.0.0](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.0)

A standalone `flashdata` executable for developers, coding agents and personal agents. [FlashData](https://flashdata.dev) provides structured Google search and YouTube data, existing transcripts/captions, background processing and customer-storage media delivery. This CLI is built on the Go SDK and adds secure profiles, recoverable onboarding, durable run IDs and streaming bulk exports.

**Version 1.0.0.** Download ready-to-run executables from the [official v1.0.0 release](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.0). The five supported targets are Linux amd64/arm64, macOS Intel/Apple Silicon, and Windows amd64. Installed binaries require no Node.js, Python, npm, Docker or separate Go installation. Source development requires Go 1.25+.

## Install v1.0.0

### Python installers (local candidate)

**PyPI publication is pending.** Local platform wheels support `pip`, `uv pip` and `uv tool`, using the same verified CLI executable. Python 3.10+ is required for this installation route; no Go installation is needed. Build instructions and supported platforms are in the [English guide](docs/python-packaging.md) and [简体中文](docs/python-packaging.zh-CN.md).

After building `dist/wheels`, choose one method from the checkout root. Use an active virtual environment for pip/uv pip:

```bash
python -m pip install --no-index --find-links dist/wheels flashdata-cli==1.0.0
# Or, inside a virtual environment:
uv pip install --no-index --find-links dist/wheels flashdata-cli==1.0.0
# Or, as an isolated CLI tool:
uv tool install --no-index --find-links dist/wheels flashdata-cli==1.0.0
```

The installed command is `flashdata`. See [installation, reinstall and removal](docs/installation.md#python-installers-local-candidate). The distribution `flashdata-cli` is separate from the Python SDK `flashdata`; direct registry installation will be available after an authorized PyPI release.

[Native CI validation](docs/python-packaging.md#validate) passed all 30 installer/platform/Python combinations. Candidate wheels are available as the `wheel-candidates` artifact in the successful GitHub Actions run; PyPI publication remains pending.

### Homebrew (Linux and macOS)

With [Homebrew](https://brew.sh) installed and on PATH:

```bash
brew install flashdatadev/tap/flashdata
```

The [official FlashData tap](https://github.com/flashdatadev/homebrew-tap) selects Linux amd64/arm64 or macOS Intel/Apple Silicon and verifies the release archive's SHA-256 checksum. No Go installation is needed. See the [installation guide](docs/installation.md#homebrew-linux-and-macos) for upgrades, uninstalling, packaged examples and switching from a manual installation.

### APT (Ubuntu and Debian)

Follow the [one-time signed repository setup](docs/installation.md#apt-ubuntu-and-debian), then install:

```bash
sudo apt-get update
sudo apt-get install flashdata
```

The [official APT repository](https://github.com/flashdatadev/apt) supports Ubuntu 22.04/24.04 and Debian 12/13 on amd64/arm64. Package **1.0.0-1** contains CLI **1.0.0**. Setup verifies the dedicated signing key fingerprint and scopes trust with `Signed-By`. The [installation guide](docs/installation.md#apt-ubuntu-and-debian) also covers upgrades, uninstalling and switching from another installation method.

### RPM (Fedora and RHEL)

Follow the [one-time signed repository setup](docs/installation.md#rpm-fedora-and-rhel), then install:

```bash
sudo dnf --refresh install flashdata
```

The [official RPM repository](https://github.com/flashdatadev/rpm) provides **1.0.0-1** for x86_64/aarch64. Native CI covers Fedora 43/44 and RHEL 8/9/10 userspace through Red Hat UBI containers. Setup verifies the dedicated RPM signing certificate's full fingerprint; both RPM packages and repository metadata are signed. See the [installation guide](docs/installation.md#rpm-fedora-and-rhel) for upgrades, trust details and switching from another installation.

### Standalone downloads

Download the archive for your operating system and CPU. No Go installation is needed.

| Platform | Download |
| --- | --- |
| Linux amd64 | [flashdata_1.0.0_linux_amd64.tar.gz](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_linux_amd64.tar.gz) |
| Linux arm64 | [flashdata_1.0.0_linux_arm64.tar.gz](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_linux_arm64.tar.gz) |
| macOS Intel | [flashdata_1.0.0_darwin_amd64.tar.gz](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_darwin_amd64.tar.gz) |
| macOS Apple Silicon | [flashdata_1.0.0_darwin_arm64.tar.gz](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_darwin_arm64.tar.gz) |
| Windows amd64 | [flashdata_1.0.0_windows_amd64.zip](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/flashdata_1.0.0_windows_amd64.zip) |

Verify the download against [SHA256SUMS](https://github.com/flashdatadev/flashdata-cli/releases/download/v1.0.0/SHA256SUMS), extract it and place the executable in a directory on PATH. The [installation guide](docs/installation.md) provides copy-and-run Linux/macOS and Windows instructions, including checksum verification and PATH setup.

```bash
flashdata --version
flashdata --help
```

The release also includes [release notes](docs/release-notes-1.0.0.md), a build report and English/Chinese READMEs.

## Build from source

Clone this version and build with Go 1.25+. The dependency is pinned to SDK **v1.0.0** and verified by `go.sum`; no sibling SDK checkout or local replacement is required. See the [v1.0.0 release notes](docs/release-notes-1.0.0.md).

```bash
git clone --branch v1.0.0 --depth 1 https://github.com/flashdatadev/flashdata-cli.git
cd flashdata-cli
go mod download
go build -trimpath -o dist/flashdata ./cmd/flashdata
./dist/flashdata --version
./dist/flashdata --help
```

Windows PowerShell:

```powershell
go build -trimpath -o dist/flashdata.exe ./cmd/flashdata
.\dist\flashdata.exe --version
.\dist\flashdata.exe --help
```

Copy the matching executable from `dist/binaries` or a locally generated archive into a user-owned directory on PATH, naming it `flashdata` (`flashdata.exe` on Windows). No service or runtime installation is required. [Release preparation](docs/releases.md) describes archives, SHA-256 checksums and OS verification; official release packages can be downloaded without signing in.

## Existing account: import a key

Obtain a key from your [FlashData account](https://flashdata.dev).

```bash
flashdata auth login
```

In a terminal the key input is hidden. CI and agents must provide one key through stdin:

```text
your-secret-manager read flashdata-api-key | flashdata auth login --api-key-stdin --non-interactive
```

The first program is illustrative: use your actual secret manager. Never pass a key as a command argument. The CLI also accepts an explicitly configured `FLASHDATA_API_KEY` for the current command; it overrides the selected saved key, is not implicitly written into the vault, and must remain outside scheduler text. `auth login` is the explicit import operation. Re-importing the same saved key preserves onboarding and upload recovery state; `--replace` explicitly discards that state. Bulk operations using an environment override preserve the original saved credentials. You can explicitly import a key into an environment-only profile without discarding accepted-job recovery. `.env` files are never loaded.

```bash
flashdata auth status
flashdata sources list
flashdata search "latest battery recycling research" --num 5 --run-id research-001
```

Normal FlashData billing applies to submitted queries and jobs. Public discovery, diagnostics and account reads do not submit paid queries.

## New agent: onboard and claim

```bash
flashdata discover
flashdata init --name "Research assistant" --client codex --scopes google_search,youtube_metadata
flashdata auth claim
# Open the returned URL, enter the separately displayed code and approve in your browser.
flashdata auth finish --wait --timeout 10m
flashdata auth status
```

`init` verifies and reuses existing access. It only creates a temporary connection when no previous profile exists and an explicit name is supplied. The registration proof and exact input are durably saved before POST; an interrupted registration recovers with that same proof. Invalid/revoked/expired credentials, a missing vault, exhausted access or a logged-out profile never trigger a replacement trial.

The starting trial allowance and request equivalents are reported separately from current balance. Per-source equivalents are alternative uses of one shared starting grant. Availability, scopes, credits and expiry come from the server.

After browser approval, `auth finish` atomically stores the candidate key, launches this executable as a fresh process to read and verify it, then sends ACK. Failed verification retains the candidate/proof and sends no ACK. After server-confirmed installation, obsolete setup/handoff/registration material is removed. Direct REST uses ordinary API keys; MCP OAuth credentials are not accepted as a substitute. See [onboarding and recovery](docs/onboarding.md).

Onboarding is feature-gated server-side. Imported-key recurring access needs the new identity endpoint from the accompanying local backend patch; 404/insufficient ownership cannot enable recurring execution. Claimed connections can use their existing authoritative connection identity endpoint. No backend deployment was performed here.

## JSON input, dry runs and output

```bash
flashdata query --input examples/query.json --dry-run
flashdata query --input examples/query.json --run-id research-002 --output research.json
flashdata youtube metadata dQw4w9WgXcQ
flashdata youtube transcript dQw4w9WgXcQ --language en
flashdata youtube captions dQw4w9WgXcQ --language en
```

`query` accepts a file or `--input -` for stdin. Canonical input is the Data API request:

```json
{"source":"google_search","query":"latest battery recycling research","params":{"num":5}}
```

The spec's flat equivalent (`"num":5` at the top level) is also supported and normalized before hashing/submission. Unknown fields, duplicate JSON fields, conflicting flat/nested parameters and invalid source values are rejected. No request defaults are invented by normalization. Query/batch files containing storage credentials or callback URLs must have private permissions; stdin is also supported. Transcript/caption convenience commands explicitly select existing automatic/uploader tracks using `auto:true` / `auto:false` and language `en` unless changed.

`--dry-run` performs local SDK validation and live access checks, and includes available account pricing. It submits no paid query or bulk creation; it is an estimate, not an enforced spending budget. Provider URLs, timestamps, extra fields and billing usage are retained; large JSON numbers are preserved. Retrieved text is treated only as data.

Without a terminal, output defaults to a versioned JSON envelope. `--json` forces it; terminal output uses an indented envelope. Success goes to stdout, errors/diagnostics to stderr. Exit status is 0 on success, 1 on failure. `--output` refuses existing files; async submissions save an acceptance receipt unless `--wait` is selected. Large bulk receipts remain in private local storage and are exported with `--output`, rather than being embedded in one huge JSON envelope.

```json
{"schemaVersion":1,"ok":true,"runId":"research-001","cached":false,"data":{},"usage":{"credits":0.02}}
```

The credit amount above is illustrative. Actual usage comes from the API.

## Recurring work and recovery

Your host owns the schedule. Put only a profile and stable occurrence ID in the scheduled command:

```bash
flashdata query --input examples/query.json --recurring --run-id daily-research:2026-10-22:Asia-Shanghai --profile research --json --non-interactive
```

Import or finish claiming that profile first. Recurring work requires current server-verified owner access and, for claimed keys, confirmed installation. Each occurrence must have its own explicit run ID. Same profile + same ID + same normalized input returns the saved result after fresh access checks, including after a restart. Different input conflicts. Different account/origin cannot read a previous run's data. A run created with `--recurring` keeps its owner/installation requirement during later cache reads and recovery, even if that flag is omitted.

```bash
flashdata runs list
flashdata runs show research-001
flashdata runs show research-001 --result --output recovered.json
flashdata query --input examples/youtube-metadata.json --mode async --run-id metadata-001
flashdata runs recover metadata-001 --output metadata.json
flashdata jobs get job_YOUR_ID
flashdata jobs wait job_YOUR_ID --timeout 10m
flashdata jobs result job_YOUR_ID --output job-result.json
flashdata batch submit --input examples/batch.json --run-id batch-001 --wait
```

Intent is committed before submission. Lost mutation responses are uncertain; failed/uncertain runs are never automatically resubmitted. `runs recover` only reads known remote work and retrieves results; it does not create a replacement, repeat a query, upload missing input or resume a pause. Cancellation/timeout stops local waiting, not admitted server work. One-off commands can generate an ID; repeating without reusing it is an intentional new execution. Protection applies within one persistent state directory and is not cross-machine exactly-once delivery. If a malformed response still contains valid IDs, errors and history retain them as `candidateRemoteIds` with `acceptanceConfirmed:false`. Inspect them with `jobs get` or `bulk get`; automatic recovery cannot declare the entire uncertain submission successful. See [run journal](docs/runs.md).

## Complete bulk workflow

Metadata, transcripts, captions, license labels and downloads use the same durable flow. Start with a free validation check, then create/upload/start/wait/export:

```bash
flashdata bulk run --source youtube_metadata --input examples/ids.txt --run-id metadata-bulk-001 --dry-run
flashdata bulk run --source youtube_metadata --input examples/ids.txt --run-id metadata-bulk-001 --timeout 1h --output metadata-receipts.jsonl
```

For media delivery, put your real destination configuration in a private file outside the repository, using [storage.example.json](examples/storage.example.json) as a shape reference. Use a private options file containing `{"download_type":"audio","output_format":"mp3"}` for audio. Never put cloud credentials in arguments, environment-generated scheduler text, or tracked files. Unix files must be owned by you with mode 0600; Windows files need an ACL granting only your user, SYSTEM and optionally Administrators.

```bash
flashdata bulk run --source youtube_download --input examples/ids.txt --storage-config /private/flashdata/storage.json --options /private/flashdata/download-options.json --run-id audio-bulk-001 --timeout 1h --output audio-receipts.jsonl
# Recover accepted work without another creation/upload:
flashdata runs recover audio-bulk-001 --timeout 1h --output recovered-receipts.jsonl
```

Media bytes go directly to your customer bucket. JSONL rows contain per-item results/delivery receipts. A completed bulk job may include failed/invalid items: those rows are saved before a nonzero exit. Partial streams are retained. Partial or per-item-failure exports return a structured error on stderr with nonzero exit status and no success envelope on stdout. Safe recovery can finish result reads and discard spent upload capabilities after an observed start. Output files are not overwritten. Support includes S3, S3-compatible, OSS, GCS, Azure, COS, BOS and TOS destination shapes; real permissions/provider-specific settings remain server-validated.

For manual creation, the following Bash commands use checked-in, non-secret metadata examples. Replace `bulk_YOUR_ID` with the ID returned by create. Apply the options patch before upload/start:

```bash
flashdata bulk create --input - --run-id bulk-manual-001 < examples/bulk-create.json
flashdata bulk update bulk_YOUR_ID --input - < examples/bulk-update.json
flashdata bulk upload bulk_YOUR_ID --input examples/ids.txt --run-id bulk-manual-001
flashdata bulk start bulk_YOUR_ID
flashdata bulk get bulk_YOUR_ID
flashdata bulk wait bulk_YOUR_ID --timeout 1h
flashdata bulk results bulk_YOUR_ID --run-id bulk-read-001 --output receipts.jsonl
```

PowerShell equivalents for the two stdin operations (choose these instead of the Bash redirections above):

```powershell
Get-Content -Raw examples/bulk-create.json | flashdata bulk create --input - --run-id bulk-manual-001
Get-Content -Raw examples/bulk-update.json | flashdata bulk update bulk_YOUR_ID --input -
```

Optional controls: pause only while the job is running, inspect its status, and resume only if it is paused. These are separate from the completed workflow above:

```bash
flashdata bulk pause bulk_YOUR_ID
flashdata bulk get bulk_YOUR_ID
flashdata bulk resume bulk_YOUR_ID
```

Upload capabilities are stored only in the protected vault. Recovery binds an upload to the original input-file hash. Input is 1 byte–2 GiB, one video ID per line. Signed uploads expire in 15 minutes; internal results are retained for 24 hours. An expired capability needs explicit inspection/support; the CLI never creates a replacement automatically. See [bulk guide](docs/bulk.md).

## Profiles, credentials and diagnostics

Global options: `--profile`, `--state-dir`, `--credential-store auto|os|file`, `--json`, `--non-interactive`, `--timeout`, `--help`. Expert origin flags configure a new profile and cannot redirect a saved profile's credentials to another service. Default state lives outside repositories in the OS user state/config directory. Profiles accept lowercase letters, digits, dots, underscores and hyphens, beginning with a letter/digit.

Windows defaults to current-user DPAPI; macOS uses native Keychain calls without CGo; Linux uses an unlocked Secret Service without GUI prompts, or private-file mode automatically when no session bus is configured. OS-store errors never downgrade an existing vault. Explicit `--credential-store file` supports headless persistent environments; it is permission-protected plaintext encoded in an envelope, **not encryption**. See [security and storage](docs/security.md).

```bash
flashdata doctor
flashdata doctor --offline
flashdata account balance
flashdata account pricing --source google_search --request-count 100
flashdata account limits
flashdata account usage --source youtube_metadata --from 2026-10-01T00:00:00Z --to 2026-10-02T00:00:00Z
flashdata commands --json
flashdata auth logout
# Existing script state is migrated only through an explicit operation:
flashdata auth migrate --from /private/legacy/connection.json --profile migrated
```

Logout removes local credential material while retaining execution history, and reports that remote access was not revoked. Unset an environment key separately. Migration preserves the source and imports history/tombstones; legacy original IDs cannot be guessed or replayed. Missing/corrupt state is never silently reset.

## Development and verification

```bash
go test ./...
go vet ./...
# Supported hosts with a C compiler:
go test -race -short ./...
# Local archives and checksums; this script does not upload:
python scripts/build_release.py --go /path/to/go
```

Tests use mock loopback APIs and dummy keys. They exercise the compiled binary, actual fresh-process installation verification, crash-released OS locks, revocation before cached reads, uncertain submissions, schema normalization and bulk partial-result recovery. No live API key, trial, paid request, cloud write, remote release or deployment is used. [Validation evidence](docs/validation.md) records exact results and host limitations; the [local review](docs/review.md) lists the fixes made against the CLI spec and backend contracts.

[Command reference](docs/commands.md) · [Security](SECURITY.md) · [Joint release check](docs/release-readiness.md) · [Third-party notices](THIRD_PARTY_NOTICES.md) · [Contributing](CONTRIBUTING.md) · [Changelog](CHANGELOG.md) · [MIT license](LICENSE)
