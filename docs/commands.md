# Command reference

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Generated from `flashdata commands --json`. Every operational command supports the global profile, state directory, credential protection, timeout and output-mode flags. Keys are accepted only through hidden input/stdin or an explicitly configured environment variable.

## flashdata

FlashData search, YouTube and bulk data workflows

Usage: `flashdata`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--credential-store` | string | `auto` | auto, os, or file (protected persistent file) |
| `--data-origin` | string | `` | pin the data origin for a new profile |
| `--json` | bool | `false` | versioned machine-readable output |
| `--management-origin` | string | `` | pin the management origin for a new profile |
| `--non-interactive` | bool | `false` | disable prompts and browser opening |
| `--profile` | string | `default` | isolated saved connection |
| `--state-dir` | string | `` | private persistent state directory |
| `--timeout` | duration | `2m0s` | overall command deadline; accepted server work continues |
| `--version` | bool | `false` | show version offline |
| `--web-origin` | string | `` | pin the browser origin for a new profile |

## flashdata account

Read shared balance, source pricing, limits and usage

Usage: `account`

## flashdata account balance

Read account balance without submitting paid work

Usage: `balance`

## flashdata account limits

Read account limits without submitting paid work

Usage: `limits`

## flashdata account pricing

Read account pricing without submitting paid work

Usage: `pricing`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--request-count` | int | `1` | estimate this many requests |
| `--source` | string | `` | filter by source |

## flashdata account usage

Read account usage without submitting paid work

Usage: `usage`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--from` | string | `` | RFC3339 start time |
| `--source` | string | `` | filter by source |
| `--to` | string | `` | RFC3339 end time |

## flashdata auth

Manage API-key profiles and browser claiming

Usage: `auth`

## flashdata auth claim

Create the browser approval URL and separate confirmation code

Usage: `claim`

## flashdata auth finish

Save, independently verify and acknowledge the approved owner key

Usage: `finish`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--wait` | bool | `false` | wait for browser approval within --timeout |

## flashdata auth login

Import and verify an API key using hidden input or stdin

Usage: `login`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--api-key-stdin` | bool | `false` | read exactly one API key from standard input |
| `--replace` | bool | `false` | explicitly replace saved onboarding credentials, retaining history |

## flashdata auth logout

Remove local credentials and retain run history; does not revoke remote access

Usage: `logout`

## flashdata auth migrate

Explicitly migrate the version-2 script state without resetting history

Usage: `migrate`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--from` | string | `` | legacy private connection.json file |

## flashdata auth status

Verify access, scopes, expiry and claim status

Usage: `status`

## flashdata batch

Background mixed-source batch submission

Usage: `batch`

## flashdata batch submit

Validate all input and submit a batch once

Usage: `submit`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--input` | string | `` | JSON file, or - for stdin |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--wait` | bool | `false` | wait for all accepted batch jobs |

## flashdata bulk

Create, upload, control and recover durable bulk tasks

Usage: `bulk`

## flashdata bulk cancel

cancel an existing bulk job; never creates replacement work

Usage: `cancel BULK_ID`

## flashdata bulk create

Create a bulk draft once and privately retain its signed upload capability

Usage: `create`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--input` | string | `` | private CreateBulkJobRequest JSON file, or - for stdin |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |

## flashdata bulk get

get an existing bulk job; never creates replacement work

Usage: `get BULK_ID`

## flashdata bulk list

List existing remote bulk jobs

Usage: `list`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--cursor` | string | `` | opaque page cursor |
| `--limit` | int | `20` | page size, 1–50 |

## flashdata bulk pause

pause an existing bulk job; never creates replacement work

Usage: `pause BULK_ID`

## flashdata bulk results

results an existing bulk job; never creates replacement work

Usage: `results BULK_ID`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--output` | string | `` | new NDJSON receipt output file |
| `--run-id` | string | `` | local recovery record to reuse |

## flashdata bulk resume

resume an existing bulk job; never creates replacement work

Usage: `resume BULK_ID`

## flashdata bulk run

Create once, upload IDs, start, wait and save NDJSON receipts

Usage: `run`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--input` | string | `` | UTF-8 video-ID file |
| `--options` | string | `` | private shared source-options JSON file |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--source` | string | `youtube_metadata` | one of five YouTube bulk sources |
| `--storage-config` | string | `` | private customer storage JSON |

## flashdata bulk start

start an existing bulk job; never creates replacement work

Usage: `start BULK_ID`

## flashdata bulk update

update an existing bulk job; never creates replacement work

Usage: `update BULK_ID`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--input` | string | `` | private partial options/callback JSON file, or - for stdin |

## flashdata bulk upload

Explicitly upload the original ID file to a saved draft

Usage: `upload BULK_ID`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--input` | string | `` | original video-ID file |
| `--run-id` | string | `` | owning local run ID (auto-detected when unique) |

## flashdata bulk wait

wait an existing bulk job; never creates replacement work

Usage: `wait BULK_ID`

## flashdata commands

Return the command tree for agent discovery

Usage: `commands`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--help` | bool | `false` | help for commands |

## flashdata discover

Read service health, onboarding capabilities and documentation

Usage: `discover`

## flashdata doctor

Diagnose storage and access without paid queries or account mutations

Usage: `doctor`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--offline` | bool | `false` | inspect local storage only |

## flashdata help

Help about any command

Usage: `help [command]`

## flashdata init

Reuse saved access or durably register one temporary connection

Usage: `init`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--audience` | string | `coding` | coding or personal |
| `--client` | string | `flashdata-cli` | self-reported client label |
| `--name` | string | `` | connection name; required to start a new trial |
| `--scopes` | string | `google_search,youtube_metadata` | comma-separated trial source names |

## flashdata jobs

Read, wait for and recover accepted background jobs

Usage: `jobs`

## flashdata jobs get

get an accepted job; never submits work

Usage: `get JOB_ID`

## flashdata jobs result

result an accepted job; never submits work

Usage: `result JOB_ID`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--output` | string | `` | new result output file |

## flashdata jobs wait

wait an accepted job; never submits work

Usage: `wait JOB_ID`

## flashdata query

Validate and execute a Data API JSON request

Usage: `query`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--input` | string | `` | JSON file, or - for stdin |
| `--mode` | string | `realtime` | realtime or async |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--wait` | bool | `false` | wait for an accepted async job |

## flashdata runs

Inspect durable local history and recover known accepted work

Usage: `runs`

## flashdata runs list

List local execution metadata without network calls

Usage: `list`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--limit` | int | `100` | maximum records displayed; history is never evicted |

## flashdata runs recover

Read accepted remote jobs and save results; never resubmit a query or create a job

Usage: `recover RUN_ID`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--output` | string | `` | new result output file |

## flashdata runs show

Inspect a record; --result revalidates live access before reading saved data

Usage: `show RUN_ID`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--output` | string | `` | new result output file |
| `--result` | bool | `false` | read the saved result after access verification |

## flashdata search

Google Search with the normal query billing and recovery path

Usage: `search QUERY`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--gl` | string | `` | country code |
| `--hl` | string | `` | interface language |
| `--mode` | string | `realtime` | realtime or async |
| `--num` | int64 | `0` | requested number of results |
| `--output` | string | `` | new result or submission-receipt output file |
| `--page` | int64 | `0` | requested page; separately billed |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--wait` | bool | `false` | wait for an accepted async job |

## flashdata sources

Discover available, permitted and CLI-supported sources

Usage: `sources`

## flashdata sources list

List source support and selected-key permissions

Usage: `list`

## flashdata sources show

Show schema, defaults, modes, example and optional account pricing

Usage: `show SOURCE`

## flashdata youtube

YouTube search, metadata, existing captions and media delivery

Usage: `youtube`

## flashdata youtube captions

YouTube captions through the shared query workflow

Usage: `captions INPUT`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--language` | string | `en` | caption language; existing tracks only |
| `--mode` | string | `realtime` | realtime or async |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--wait` | bool | `false` | wait for an accepted async job |

## flashdata youtube download

Deliver media to private customer storage (background only)

Usage: `download VIDEO_ID`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--format` | string | `mp3` | output format |
| `--mode` | string | `realtime` | realtime or async |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--storage-config` | string | `` | private customer-storage JSON file |
| `--wait` | bool | `false` | wait for an accepted async job |

## flashdata youtube metadata

YouTube metadata through the shared query workflow

Usage: `metadata INPUT`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--mode` | string | `realtime` | realtime or async |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--wait` | bool | `false` | wait for an accepted async job |

## flashdata youtube search

YouTube search through the shared query workflow

Usage: `search INPUT`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--max-results` | int64 | `0` | requested number of search results |
| `--mode` | string | `realtime` | realtime or async |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--wait` | bool | `false` | wait for an accepted async job |

## flashdata youtube transcript

YouTube transcript through the shared query workflow

Usage: `transcript INPUT`

| Flag | Type | Default | Purpose |
| --- | --- | --- | --- |
| `--dry-run` | bool | `false` | validate without a paid submission |
| `--language` | string | `en` | caption language; existing tracks only |
| `--mode` | string | `realtime` | realtime or async |
| `--output` | string | `` | new result or submission-receipt output file |
| `--recurring` | bool | `false` | require verified owned/installed access |
| `--run-id` | string | `` | stable execution ID; required for recurring work |
| `--wait` | bool | `false` | wait for an accepted async job |
