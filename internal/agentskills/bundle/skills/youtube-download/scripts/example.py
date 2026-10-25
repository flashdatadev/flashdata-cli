import asyncio
import json
import os
from mcp_response import call_complete
from pathlib import Path

import httpx
from mcp import ClientSession
from mcp.client.streamable_http import streamable_http_client

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


async def main():
    async with httpx.AsyncClient(
        headers={"X-API-Key": os.environ["FLASHDATA_API_KEY"]},
        timeout=120, follow_redirects=False, trust_env=False,
    ) as http:
        async with streamable_http_client(
            "https://data.flashdata.dev/mcp", http_client=http,
        ) as (read, write, _):
            async with ClientSession(read, write) as session:
                await session.initialize()
                data = await call_complete(session, "youtube_download", arguments)
                if not isinstance(data, dict) or not data.get("id"):
                    raise RuntimeError("Missing job ID; check Jobs before resubmitting.")
                data = {key: data[key] for key in ("id", "status", "source", "created_at") if key in data}
                print(json.dumps(data, ensure_ascii=False, indent=2))


try:
    asyncio.run(main())
except Exception:
    raise SystemExit("Request failed. Check Jobs before retrying a paid submission; no automatic retry was made.") from None
