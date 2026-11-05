# Video research recipe

Use this when the user wants evidence from YouTube about a topic, a comparison of videos, or analysis of specific video links.

Choose the depth from the question. A title/description comparison may need only search results; claims about what a speaker said require transcript evidence. If the user supplies video links, reuse their IDs instead of buying a discovery search.

## Retrieve only what is needed

For discovery, `youtube_search` accepts `query` and `max_results`. Set an explicit small result limit appropriate to the user's request; the current server default can return 20. Optional `filters` can restrict upload date, duration, type, or sort order. Follow the discovered schema.

Select videos by relevance, source, date, and available evidence, explaining material gaps. Reuse returned `video_id` values or IDs from the user's links; never invent IDs. `youtube_metadata` accepts `video_id`, `include_formats`, and `include_chapters`; omit unnecessary format details by setting `include_formats: false` when supported.

`youtube_transcript` accepts `video_id`, `language`, and `auto`:

- `language` may be one code or a list of 1–10 distinct codes. Omit it to prefer `en-orig`, then `en`, returning one track. Preserve explicit language choices.
- `auto: true` selects automatically generated captions; `false` selects manually supplied captions. This flag does not fall back between the two.
- If the requested transcript is unavailable, say so. Try another language or caption type only when it fits the user's existing scope and budget; each attempt can be another paid call.

For a request involving one search and up to three videos, a useful cost estimate is one `youtube_search` call plus the actually needed metadata/transcript calls, each quoted separately with `get_pricing`. Three search results do not mean three search calls.

## Deliver evidence

For each analyzed video, retain its source URL, title, and observed publication information when available. Mark whether a finding comes from metadata, the description, or an actual transcript. Reference transcript start times, for example a `t=` link at a supplied timestamp, without fabricating timing or a verbatim quote. Say when automatic captions make a passage uncertain.

Compare supported claims, disagreements, and unanswered questions. A video lacking transcripts can still be listed as a discovery result, but do not present its unseen content as reviewed evidence. Retrieving a transcript does not authorize a media download or a downstream publication.

For a language list, read `results[0].tracks` and `results[0].errors`. Each track
has the usual text, language and timestamp fields. Each error has `languageCode`,
`code` (`language_unavailable` or `retrieval_failed`) and `message`. Preserve available
tracks and report missing languages; both lists follow request order. If all languages
fail, `tracks` is empty. Video-level extraction failures still fail the request.
Do not resubmit the whole query just because one language failed.
