# Runnable inputs

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Use query.json with query --input, query-flat.json for the spec-compatible flat form, youtube-metadata.json for async recovery, batch.json for batch submit, and ids.txt for bulk data jobs. These files contain public query/video examples, not credentials.

bulk-create.json and bulk-update.json are runnable metadata inputs for manual creation and a pre-upload options patch. Feed them through --input - using Bash redirection or PowerShell Get-Content -Raw; replace the remote bulk ID in later commands. Metadata patches after a scan has completed are restricted by the backend; use only the credential-only recovery patch documented for an eligible paused customer-delivery job.

bulk-options.json and download-options.json show option shapes. Copy a chosen option into a private file before passing --options. Copy storage.example.json to a private location outside the checkout and replace every placeholder with your actual authorized destination configuration. The CLI requires private storage/config permissions; Unix example files are intentionally not treated as a secret vault.

Offline command and compiled-binary tests cover the onboarding, query, batch, bulk and recovery workflows, using generated private fixture storage instead of real cloud credentials. Tests also validate example JSON and check that both README translations use identical command blocks. Secret-manager names, private paths, remote IDs and storage values in the docs are placeholders to replace. Connecting these examples to production incurs ordinary API charges; a dry-run validates and reads metadata without a paid submission.

## Subtitle languages

```sh
flashdata youtube transcript aircAruvnKk --accept text/plain --json=false
flashdata youtube captions aircAruvnKk --language en
flashdata youtube transcript aircAruvnKk --language en,es --accept application/json --json
```

Omitting `--language` prefers `en-orig`, then `en`, returning one track. A single code
selects that exact language and keeps the original single-track response. Supply
1–10 distinct codes, separated by commas or repeated flags, for multiple languages.
Each code must match `[A-Za-z0-9-]{1,35}`. The transcript command selects automatic
captions; the captions command selects uploaded captions. Neither switches source
or translates missing tracks. Use `query --input` to choose `params.auto` explicitly.

A language list returns `data.results[0].tracks` and `data.results[0].errors` in the
CLI's JSON envelope. Preserve available tracks and inspect every error's
`languageCode`, `code` (`language_unavailable` or `retrieval_failed`) and `message`.
Both lists follow request order. If all languages fail, `tracks` is empty; video-level
extraction failures still fail the request. Do not resubmit the whole request for one missing track.

The equivalent JSON input is:

```json
{"source":"youtube_transcript","video_id":"aircAruvnKk","params":{"language":["en","es"],"auto":true}}
```

This works with `query`, async queries and `batch submit`. A one-element JSON array
still returns `tracks` and `errors`; use `query --input` when you need that shape.
`Accept` and `X-Respond-With` choices continue to apply independently of language.
For bulk jobs, put `language` and `auto` in the shared options file; exports stay JSONL.
