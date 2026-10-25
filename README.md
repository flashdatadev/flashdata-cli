# FlashData CLI

[简体中文](README.zh-CN.md) · [API docs](https://flashdata.dev/docs) · [Pricing](https://flashdata.dev/pricing)

Google Search and YouTube data for your terminal and AI agent. Search the web, retrieve video metadata and transcripts, or process video lists with bulk jobs.

## Get started

Install with [uv](https://docs.astral.sh/uv/) and set up your agent:

```sh
uv tool install flashdata-cli
flashdata setup
```

Choose Codex, Claude Code or Cursor and confirm the project directory. Setup installs the skills and configures access without a paid query.

**Try before signing up.** Eligible new users can start with temporary access and limited Credits for Google Search and YouTube metadata. Existing credentials are reused. Already have a key? Run `flashdata auth login` before setup. See [account setup](docs/onboarding.md) for trial limits and signing up later.

Run your first search:

```bash
flashdata search "battery recycling startups" --num 5
```

For terminal-only use, run `flashdata init --name "My CLI"` instead of `setup`. Data queries use Credits; `flashdata account balance` shows your remaining allowance.

<details>
<summary>Other installation methods</summary>

With Python 3.10+ in an active virtual environment, choose pip or uv pip:

```sh
python -m pip install --upgrade flashdata-cli
```

```sh
uv pip install --upgrade flashdata-cli
```

Or use Homebrew on macOS/Linux:

```sh
brew install flashdatadev/tap/flashdata
```

Then run `flashdata setup`. [APT, RPM and standalone downloads](docs/installation.md) are also available. Native packages need no Python, Go or Node.js.

</details>

## Common commands

Save Google search results:

```bash
flashdata search "battery recycling startups" --num 5 --output search.json
```

Find YouTube videos, then use a returned video ID to get metadata or an existing transcript:

```bash
flashdata youtube search "Go programming tutorial" --max-results 5
flashdata youtube metadata dQw4w9WgXcQ --output video.json
flashdata youtube transcript dQw4w9WgXcQ --language en --output transcript.json
```

Replace the example video ID with your own. Transcript availability depends on the video's language tracks. Use `youtube captions` for uploader captions.

For Google news, images, shopping, videos, places and other YouTube sources, run `flashdata sources list`, then `flashdata sources show SOURCE`. Submit the documented input with `flashdata query --input query.json`. See the [command reference](docs/commands.md) and [request examples](examples/README.md).

## Output formats

CLI queries return JSON by default. Choose plain text or Markdown when needed:

```bash
flashdata search "vector databases" --accept text/plain
flashdata search "vector databases" --accept text/plain --respond-with markdown
```

`--output FILE` saves the selected response body to a new file. `--json` wraps text in the CLI's JSON envelope. [Agent skills](https://github.com/flashdatadev/flashdata-skills) request plain text by default; bulk exports always use JSONL. See [response formats](docs/response-formats.md) for async jobs and batches.

<a id="complete-bulk-workflow"></a>

## Bulk jobs

Create `ids.txt` with one YouTube video ID per line:

```text
dQw4w9WgXcQ
jNQXAC9IVRw
```

Export metadata for the list:

```bash
flashdata bulk run --source youtube_metadata --input ids.txt --run-id video-dataset-001 --timeout 1h --output videos.jsonl
```

The command submits the job, waits and saves one JSONL row per video. Bulk jobs also support transcripts, captions, license classifications and downloads. Downloads deliver media to your cloud storage; the JSONL export contains delivery receipts.

If interrupted, recover the same job into a new file:

```bash
flashdata runs recover video-dataset-001 --timeout 1h --output recovered-videos.jsonl
```

Recovery reads existing work without submitting another job. See the [bulk guide](docs/bulk.md) for MP3/MP4 examples, storage setup, failed items and job controls.

## Upgrade

```sh
uv tool upgrade flashdata-cli
flashdata setup
```

Setup refreshes unmodified skills and preserves local edits. For other installers, follow the [upgrade guide](docs/installation.md).

## More

- [Agent setup](docs/agent-setup.md): select skills, install globally, or use `--skip-auth` to install skills offline without changing credentials. MCP is optional.
- [Account setup](docs/onboarding.md): temporary access, API keys and connecting your account.
- [Request examples](examples/README.md): background queries, batches and full source options.
- [Credentials and automation](docs/security.md) · [Run recovery](docs/runs.md) · [Contributing](CONTRIBUTING.md)

[Report an issue](https://github.com/flashdatadev/flashdata-cli/issues) · [Releases](https://github.com/flashdatadev/flashdata-cli/releases) · [MIT license](LICENSE)
