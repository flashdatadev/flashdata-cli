# FlashData CLI

[简体中文](README.zh-CN.md) · [FlashData](https://flashdata.dev) · [API docs](https://flashdata.dev/docs) · [Pricing](https://flashdata.dev/pricing)

**Google search and YouTube data, straight from your terminal.** [FlashData](https://flashdata.dev) gives you structured search results, video metadata and existing transcripts through one API. Use `flashdata` to research a topic, feed web results to an AI assistant, build a video dataset, or deliver audio to your cloud storage.

No application code is required. Commands return JSON by default, with text and Markdown options. Bulk jobs export JSONL, and saved run IDs let scripts recover accepted work after an interruption.

## Choose a workflow

| What you want to build | Start here |
| --- | --- |
| Web context and source links for an AI assistant | [Google search](#search-the-web-for-research-or-ai-context) |
| Industry research or news monitoring | [News search](#search-the-web-for-research-or-ai-context) |
| Video discovery for a topic or learning resource | [YouTube search](#find-youtube-videos-and-collect-their-data) |
| Video analytics and catalog enrichment | [Video metadata](#find-youtube-videos-and-collect-their-data) |
| Text for video summaries, search or retrieval | [Existing transcripts](#find-youtube-videos-and-collect-their-data) |
| A dataset from a list of video IDs | [Bulk metadata export](#build-a-video-dataset-with-one-bulk-command) |
| Audio or video files delivered to your own cloud bucket | [Bulk media download](#complete-bulk-workflow) |
| Continue an interrupted bulk job | [Recovery](#recover-interrupted-work) |

The examples collect data; your application provides summarization, indexing, scheduling
and AI model calls.

## Get started

### Install

Choose one installation method. **[Homebrew](https://brew.sh)** works on macOS and Linux:

```bash
brew install flashdatadev/tap/flashdata
```

**pip** and **uv pip** work on Linux, macOS and Windows with Python 3.10+. In an active virtual environment, install from [PyPI](https://pypi.org/project/flashdata-cli/) using pip:

```bash
python -m pip install --upgrade flashdata-cli
```

Or **uv pip**:

```bash
uv pip install --upgrade flashdata-cli
```

Both install the `flashdata` command. See [Python installation](docs/installation.md#python-pip-and-uv) for virtual environment setup and supported platforms.

Other options include [APT for Ubuntu/Debian](docs/installation.md#apt-ubuntu-and-debian), [RPM for Fedora/RHEL](docs/installation.md#rpm-fedora-and-rhel), and [standalone downloads for Linux, macOS and Windows](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.1.2). These install the executable directly and need no Python, Go or Node.js.

### Make your first search

Initialize the CLI, then run a search:

```bash
flashdata init --name "my-cli"
flashdata search "battery recycling startups" --num 5 --json
```

`init` reuses configured access or creates and saves temporary trial access for a new profile. The default trial scopes are Google Search and YouTube metadata. To continue with your FlashData account or choose other trial sources, see [onboarding](docs/onboarding.md). If you already have an API key, use `flashdata auth login` instead and paste it at the hidden prompt.

On a fresh headless Linux setup, `flashdata auth login` automatically uses a private credential file if the desktop keyring is unavailable. Later commands reuse it without extra flags. The file has owner-only permissions and is not encrypted; see [credential storage](docs/security.md).

The search returns page titles, links and snippets. Queries use your available credits; see [pricing](https://flashdata.dev/pricing).

## JSON, plain text and Markdown

Version 1.1.2 includes `--accept` and `--respond-with` with every installation
method above. To install from source with Go 1.25+:

```bash
go install github.com/flashdatadev/flashdata-cli/cmd/flashdata@v1.1.2
```

Choose plain text or Markdown for a search:

```bash
flashdata search "vector databases" --accept text/plain
flashdata search "vector databases" --accept text/plain --respond-with markdown
```

JSON remains the default. `--accept` also applies to async receipts, batches and
job status. Markdown applies to realtime queries and completed results. Add
`--json` to wrap text in the CLI envelope, or `--output FILE` to save the selected
body. See [response formats and recovery](docs/response-formats.md).

## Search the web for research or AI context

Save search results to a file for your script or assistant:

```bash
flashdata search "battery recycling startups" --num 5 --output search.json
```

Use `--gl us` and `--hl en` to set the search country and interface language. For news, images, shopping, videos or places, use `query` with the matching source. For example, save this as `news-query.json`:

```json
{"source":"google_news","query":"battery recycling"}
```

Then run:

```bash
flashdata query --input news-query.json --output news.json
```

See [available sources and inputs](https://flashdata.dev/docs) or run `flashdata sources list` and `flashdata sources show google_news`.

## Find YouTube videos and collect their data

**Find videos about a topic:**

```bash
flashdata youtube search "Go programming tutorial" --max-results 5 --json
```

**Get a video's title, channel, duration and engagement data:**

```bash
flashdata youtube metadata dQw4w9WgXcQ --output video.json
```

**Get an existing English transcript for summarization or analysis:**

```bash
flashdata youtube transcript dQw4w9WgXcQ --language en --output transcript.json
```

Use a video ID from the search results in place of `dQw4w9WgXcQ`. Transcript availability depends on the video's existing language tracks; this does not transcribe arbitrary audio. Use `youtube captions` to request an existing manual caption track instead.

**Read an existing uploaded caption track:**

```bash
flashdata youtube captions dQw4w9WgXcQ --language en --output captions.json
```

`--json` writes a machine-readable response to stdout; `--output FILE` saves the result. Choose a new output filename for each export: existing files are never overwritten.

## Build a video dataset with one bulk command

Create `ids.txt` with one YouTube video ID per line:

```text
dQw4w9WgXcQ
jNQXAC9IVRw
```

Collect metadata for all IDs and export the per-video results:

```bash
flashdata bulk run --source youtube_metadata --input ids.txt --run-id video-dataset-001 --timeout 1h --output videos.jsonl
```

This command creates the job, uploads your IDs, starts processing, waits and writes `videos.jsonl`. Each line records one video's outcome and data. Change `--source` to `youtube_transcript`, `youtube_captions` or `youtube_trainability` for other bulk datasets. See the [bulk guide](docs/bulk.md) for options and item statuses.

## Complete bulk workflow

### Deliver audio files to your cloud bucket

Use the same `ids.txt` to request MP3 audio for multiple videos. Media files are delivered **to your bucket**; the local JSONL file contains delivery receipts.

**1. Configure your destination.** In a private working directory outside source control, create `storage.json` and replace every placeholder with your own S3 bucket and credentials:

```json
{
  "storage_type": "s3",
  "storage_url": "YOUR_BUCKET_NAME",
  "region": "us-east-1",
  "access_key": "YOUR_STORAGE_ACCESS_KEY",
  "secret_key": "YOUR_STORAGE_SECRET_KEY",
  "object_prefix": "flashdata/audio"
}
```

Create `audio-options.json` in the same directory:

```json
{"download_type":"audio","output_format":"mp3"}
```

On Linux/macOS, make both files private before running the command:

```bash
chmod 600 storage.json audio-options.json
```

On Windows, restrict both files' ACLs to your user, SYSTEM and optionally Administrators; see [private-file requirements](docs/security.md). Other supported destinations include S3-compatible storage, OSS, GCS, Azure, COS, BOS and TOS; see [storage options](docs/bulk.md).

**2. Run the complete task.** Run from the directory containing those files and `ids.txt`:

```bash
flashdata bulk run --source youtube_download --input ids.txt --storage-config storage.json --options audio-options.json --run-id audio-bulk-001 --timeout 1h --output audio-receipts.jsonl
```

The command creates, uploads, starts, waits and exports. Check `audio-receipts.jsonl` for each video's delivery status and destination. A completed job can still contain failed items; those receipts are saved and the command exits with a nonzero status.

**For one video**, use the same storage configuration:

```bash
flashdata youtube download dQw4w9WgXcQ --storage-config storage.json --format mp3 --wait --timeout 30m --output audio-delivery.json
```

### Deliver MP4 video to your cloud bucket

Use the same `ids.txt` and private `storage.json`, and create `video-options.json`:

```json
{"download_type":"audio_video","video_quality":"<=720","output_format":"mp4"}
```

On Linux/macOS, restrict the options file first:

```bash
chmod 600 video-options.json
```

```bash
flashdata bulk run --source youtube_download --input ids.txt --storage-config storage.json --options video-options.json --run-id video-bulk-001 --timeout 1h --output video-receipts.jsonl
```

Video files go to the bucket; per-video delivery outcomes go to `video-receipts.jsonl`.
On Windows, apply the same private ACL requirements as for the audio options file.

### Recover interrupted work

Keep the same profile and state directory. List saved runs, then recover the accepted bulk job into a fresh output file:

```bash
flashdata runs list
flashdata runs recover audio-bulk-001 --timeout 1h --output recovered-receipts.jsonl
```

Recovery reads the existing job; it does not submit another download. For the metadata dataset above, substitute `video-dataset-001`; for MP4 delivery, use `video-bulk-001`. If an interruption occurred before upload/start, inspect the run and follow the [bulk recovery guide](docs/bulk.md). A timeout stops local waiting; accepted server work continues.

Use a new run ID only when you intend to create a new task. See [run history and recovery](docs/runs.md) for scripts, uncertain submissions and partial exports.

## Supported APIs

| Capability | Commands |
| --- | --- |
| Google search, images, news, shopping, videos and places | `search`; `query --input FILE` for all six sources |
| All nine YouTube sources | `youtube search/metadata/transcript/captions/download`; `query` for extended search, channel, suggestions, license labels and full download inputs |
| Realtime, background and mixed batches | `query`, `query --mode async`, `batch submit` |
| Standard jobs | `jobs get`, `jobs wait`, `jobs result` |
| All five bulk sources and lifecycle/results | `bulk run`, manual `bulk` commands, `runs recover` |
| Access, balance, prices, limits and usage | `auth status`, `account balance/pricing/limits/usage` |
| Public source catalog and health | `sources list`, `discover` |
| Public plan catalog and status | `discover plans`, `discover status` |

All commands above are included in the ready-to-run **v1.1.2** release. `discover plans` and `discover status` need no API key. Upgrade an existing installation with `brew update && brew upgrade flashdatadev/tap/flashdata`, `sudo apt-get update && sudo apt-get install --only-upgrade flashdata`, or `sudo dnf --refresh upgrade flashdata`; see the [installation guide](docs/installation.md) for repository setup and Windows upgrades. Your profiles and run history are preserved.

All 15 sources support background/batch execution; 14 support realtime. Downloads run in the background and require your cloud storage. Use `sources show SOURCE` for source schemas and a runnable JSON request. The single-video `youtube download` shortcut delivers audio; `query --mode async` and bulk expose the full media source options.

See the [SDK and CLI comparison](docs/sdk-parity.md) and [command reference](docs/commands.md).

## More workflows

**Check before spending credits:** `--dry-run` validates input without submitting a paid query. Balance reads are also free.

```bash
flashdata search "battery recycling startups" --num 5 --dry-run
flashdata account balance
flashdata --help
```

- **Scripts and AI agents:** use `--json --non-interactive`, [profiles and credentials](docs/security.md), and [stable run IDs](docs/runs.md). Store a key through `auth login --api-key-stdin` when integrating a secret manager, or keep `FLASHDATA_API_KEY` set for each command to use a key without saving it. The environment key overrides any saved key; `.env` files are not loaded.
- **Background queries and mixed batches:** see the [runnable inputs](examples/README.md) and [command reference](docs/commands.md).
- **Trial setup and connecting your account:** see [onboarding](docs/onboarding.md) for `init`, browser approval and recovery.
- **Use FlashData inside a Go application:** install the [Go SDK](https://github.com/flashdatadev/flashdata-go), which powers this CLI.

[Installation and upgrades](docs/installation.md) · [API documentation](https://flashdata.dev/docs) · [Troubleshooting](docs/security.md) · [Contributing and source builds](CONTRIBUTING.md) · [Release notes](docs/release-notes-1.1.2.md) · [MIT license](LICENSE)
