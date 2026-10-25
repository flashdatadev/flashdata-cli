# Optional MCP and Python usage

Use this guide only when the user chooses MCP or a Python integration. The normal skill workflow uses the standalone FlashData CLI.

# YouTube Video and Audio Download

Send a YouTube video, audio track or selected clip directly to your own cloud storage. Choose the media type, format and quality, then use the returned job ID to retrieve the completed file location and size.

[Documentation](https://flashdata.dev/docs/sources/youtube_download?utm_source=skills-sh&utm_medium=integration&utm_campaign=intent_gtm_2026q4&utm_content=youtube_download-task) · [Get an API key](https://flashdata.dev/app/api-keys)

## Example requests

- “Save the first 30 seconds of this video to my S3 bucket.”
- “Download this video’s audio as MP3 to my storage.”
- “Deliver a 720p MP4 and its metadata to my media pipeline.”

## Quick start with Python: a clip to your storage

Use Python 3.10+ and a [FlashData API key](https://flashdata.dev/app/api-keys) supplied as `FLASHDATA_API_KEY` through your environment or secret manager. Install the Python dependencies in your virtual environment:

```sh
python3 -m pip install 'mcp==1.30.0' 'httpx==0.28.1'
```

Use the bundled [example script](../scripts/example.py). It connects to hosted [FlashData MCP](https://flashdata.dev/mcp) and sends one request. Data calls use paid [FlashData Credits](https://flashdata.dev/pricing); installing the dependencies is free. Run only within the user's authorized task and budget.

Set `FLASHDATA_STORAGE_FILE` to a private JSON file containing your destination configuration; [storage setup](storage.md) shows its fields. This example delivers the first 10 seconds as MP4. Replace the video and clip bounds for the user’s request. The file is read locally and its credentials are sent to FlashData for the upload, but are not printed.

Edit the inputs in [example.py](../scripts/example.py) for your task:

```python
arguments = {
    "video_id": "jNQXAC9IVRw",
    "mode": "async",
    "download_type": "audio_video",
    "video_quality": "<=720",
    "output_format": "mp4",
    "start_at": "00:00:00",
    "end_at": "00:00:10",
    "storage": json.loads(Path(os.environ['FLASHDATA_STORAGE_FILE']).read_text(encoding='utf-8')),
}
```

Run `python3 scripts/example.py` from this skill directory. Keep [mcp_response.py](../scripts/mcp_response.py) beside the script; it retrieves all result chunks before closing the session.

With [FlashData MCP](https://flashdata.dev/mcp), call `youtube_download` using the same fields and a `storage` object from the user's protected configuration.

## Choose the media

| Parameter | Use |
| --- | --- |
| `video_id` | Required 11-character YouTube video ID. |
| `download_type` | `audio_video` (default), `video` or `audio`. |
| `video_quality` | `<=720` by default; also `best`, `worst`, or `<=`, `==`, `>=` with 144, 360, 480, 720, 1080, 1440, 2160 or 4320. |
| `output_format` | Video: `mp4` (default), `webm`, `mkv`. Audio: `m4a` (default), `mp3`, `opus`. |
| `start_at`, `end_at` | Clip times as `HH:MM:SS`; end must follow start. Omit for the full video. |
| `storage` | Required destination configuration; see [storage setup](storage.md). |
| `mode` | `async`; downloads return a job ID. |

### Audio as MP3

Use `download_type: "audio"`, `output_format: "mp3"` and optionally `audio_quality: "192"`. Audio quality accepts `best` (default), `320`, `256`, `192` or `128`. Omit `video_quality` and `fps` for audio-only requests.

### Higher resolution or a metadata sidecar

Set `video_quality: "<=1080"` for a maximum of 1080p. Add `upload_metadata: true` to upload a metadata JSON sidecar alongside the media. Availability depends on the video's formats.

Advanced selection supports `fps` constraints such as `<=30`, `audio_language: "default"` or `"original"`, and `yt_dlp_format`. A custom `yt_dlp_format` overrides `video_quality`, `fps` and `audio_language`. Use `audio_quality` only with audio downloads and omit `audio_language` for video-only downloads.

## Read the job and get the file

The submission response contains `id`. Set `FLASHDATA_JOB_ID` to that returned ID. Run the bundled [job reader](../scripts/read_job.py) to read the existing job once:

Edit the inputs in [read_job.py](../scripts/read_job.py) for your task:

```python
arguments = {
    "job_id": os.environ['FLASHDATA_JOB_ID'],
}
```

Run `python3 scripts/read_job.py` from this skill directory. Keep [mcp_response.py](../scripts/mcp_response.py) beside the script; it retrieves all result chunks before closing the session.

While `job.status` is `pending` or `processing`, read the same job again as needed. When it is `completed`, inspect `output.results[0]` for `file_url`, `file_size`, `downloaded_bytes`, `md5` and `output_format`. A private storage URL requires access through your storage provider.

Downloads reserve Credits before work and settle by actual uploaded bytes. `get_pricing` with `{"source":"youtube_download"}` returns the per-TB rate; it cannot give an exact file cost from duration alone. Job reads are free. Use this for media the user is authorized to download.

## Connection and errors

For hosted MCP, connect `https://data.flashdata.dev/mcp` using Streamable HTTP and an `X-API-Key` header in your client's protected settings. Keep keys out of chat and committed files.

Storage credentials are transmitted to FlashData to perform the upload; MCP clients may retain tool arguments in their history. Use narrowly scoped credentials, and do not print credentials or signed access tokens in the returned file summary.

If a submission times out, do not resubmit automatically; it may already have been charged. Use `get_job` if an ID was returned. Without an ID, report the outcome as unknown and check Jobs in the FlashData console; `get_job` cannot look up a job without its ID. A failed or pending job is not a completed upload. For a chunked MCP response, use `get_result` with the returned `result_id` and `next_offset` as `offset`, then join the text chunks in order. Parse only `encoding: "json"`. The Python examples retrieve all chunks automatically in the same session.

## Output format

For agents, prefer plain text. Select `response_format: "text"` when the connected schema supports it. Local MCP 0.3.0 returns Markdown text by default and retains original fields in `structuredContent`. Use `response_format: "text"` or `"json"` when needed; `get_result` inherits the original format. Check the connected server’s schema before using this option with hosted or older servers.

The bundled Python examples request JSON when supported so programs can read the fields described above. They also work with older JSON-only servers, finish chunk reads before closing, and never retry the original paid call.
