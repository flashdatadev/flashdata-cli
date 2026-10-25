# Knowledge preparation recipe

Use this when the user wants retrieved web-search or YouTube material organized for later research, retrieval, or a knowledge base.

Reuse available results before submitting more paid work. Determine whether the intended records need search snippets, metadata, transcripts, or summaries. FlashData does not extract the full text of arbitrary webpages; do not silently turn a search snippet into a full-document record. Use another user-authorized retrieval method if the task genuinely requires webpage bodies.

## Prepare traceable records

Follow the user's destination schema when supplied. Otherwise, a practical local Markdown or JSON record contains:

- Stable source URL and a source ID when returned (such as `video_id`).
- Title, creator/publisher, language, and publication date where observed.
- Retrieval time and evidence type (`search_snippet`, `video_metadata`, or `transcript`).
- A concise summary, relevant supporting excerpts, and source links.
- Transcript segment start times when present; retain original timing when combining segments.
- Coverage notes, such as missing captions or an unavailable requested language.

Unknown fields stay unknown. Keep generated summaries distinguishable from source text. Ingest only the amount needed for the user's task and rights; do not infer reuse permissions from successful retrieval or a video's license-classification tool.

Deduplicate by stable source identity. A new retrieval time alone does not make a new document. If the user requests incremental updates, compare existing content and update the appropriate record rather than appending duplicates.

## Deliver or ingest

Local files are a usable deliverable when no destination integration is available. If the user requested a specific remote knowledge base, use its actual connected write tool and the authorized collection. Clarify a missing target or replacement policy when needed; preparing records does not authorize overwriting unrelated content or publishing it publicly.

Report how many records were actually prepared and, separately, how many the destination confirms as written. Retain destination IDs needed for future updates. A prepared JSON file, an attempted upload, and a searchable knowledge-base entry are different completion states.
