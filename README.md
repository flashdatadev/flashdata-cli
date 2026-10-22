# FlashData CLI

[简体中文](README.zh-CN.md) · [FlashData](https://flashdata.dev) · [API docs](https://flashdata.dev/docs) · [Pricing](https://flashdata.dev/pricing)

**Google search and YouTube data, straight from your terminal.** [FlashData](https://flashdata.dev) gives you structured search results, video metadata and existing transcripts through one API. Use `flashdata` to research a topic, feed web results to an AI assistant, build a video dataset, or deliver audio to your cloud storage.

No application code is required. Commands return JSON, bulk jobs export JSONL, and saved run IDs let scripts recover accepted work after an interruption.

## Get started

On macOS or Linux with [Homebrew](https://brew.sh):

```bash
brew install flashdatadev/tap/flashdata
```

For **Ubuntu/Debian**, **Fedora/RHEL** or **Windows**, follow the [installation guide](docs/installation.md), or download a [standalone executable](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.0.0). Native installs need no Go, Python or Node.js. Linux and macOS support Intel/AMD and ARM64; Windows supports x64.

Get an API key from your [FlashData account](https://flashdata.dev). Run `auth login` and paste it at the hidden prompt, then make your first search:

```bash
flashdata auth login
flashdata search "battery recycling startups" --num 5 --json
```

The result includes structured Google search data such as page titles, links and snippets. Your key is saved for later commands using the platform credential store where available; see [credential storage](docs/security.md). Queries and jobs use your account's credits; see [pricing](https://flashdata.dev/pricing).

Looking for `pip` or `uv`? The Python installer package is **not published to PyPI yet**. Use the install methods above, or follow the [candidate-wheel guide](docs/python-packaging.md).

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

### Recover interrupted work

Keep the same profile and state directory. List saved runs, then recover the accepted bulk job into a fresh output file:

```bash
flashdata runs list
flashdata runs recover audio-bulk-001 --timeout 1h --output recovered-receipts.jsonl
```

Recovery reads the existing job; it does not submit another download. For the metadata dataset above, substitute `video-dataset-001`. If an interruption occurred before upload/start, inspect the run and follow the [bulk recovery guide](docs/bulk.md). A timeout stops local waiting; accepted server work continues.

Use a new run ID only when you intend to create a new task. See [run history and recovery](docs/runs.md) for scripts, uncertain submissions and partial exports.

## More workflows

**Check before spending credits:** `--dry-run` validates input without submitting a paid query. Balance reads are also free.

```bash
flashdata search "battery recycling startups" --num 5 --dry-run
flashdata account balance
flashdata --help
```

- **Scripts and AI agents:** use `--json --non-interactive`, [profiles and credentials](docs/security.md), and [stable run IDs](docs/runs.md). Store a key through `auth login --api-key-stdin` when integrating a secret manager. An explicitly set `FLASHDATA_API_KEY` overrides the saved key for that command; `.env` files are not loaded.
- **Background queries and mixed batches:** see the [runnable inputs](examples/README.md) and [command reference](docs/commands.md).
- **New agent setup and browser claiming:** see [onboarding](docs/onboarding.md). This optional flow depends on server availability; existing accounts can use `auth login` directly.
- **Use FlashData inside a Go application:** install the [Go SDK](https://github.com/flashdatadev/flashdata-go), which powers this CLI.

[Installation and upgrades](docs/installation.md) · [API documentation](https://flashdata.dev/docs) · [Troubleshooting](docs/security.md) · [Contributing and source builds](CONTRIBUTING.md) · [Release notes](docs/release-notes-1.0.0.md) · [MIT license](LICENSE)
