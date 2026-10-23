# Onboarding and recovery

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

## Start a trial

Initialize the CLI and make your first search:

```bash
flashdata init --name "my-cli"
flashdata search "battery recycling startups" --num 5 --json
```

`init` verifies and reuses configured access. For a new profile without credentials, it creates temporary trial access and saves it locally. The default trial scopes are `google_search` and `youtube_metadata`.

Before your first `init`, you can add `--scopes youtube_search,youtube_metadata,youtube_transcript` to request different trial sources. Passing `--scopes` when reusing existing access does not expand its permissions.

The response shows approved sources, the starting allowance and available balance. Per-source request counts are alternative uses of one shared credit allowance. They are not separate quotas.

Run `flashdata discover --json` to check current trial availability and limits. If trial registration is unavailable, use an existing API key as described below.

## Connect your FlashData account

To continue beyond the trial, connect this profile to your FlashData account:

```bash
flashdata auth claim
```

Open the displayed URL, sign in or create an account, confirm the displayed code and review the requested access. Complete the browser approval, then return to the terminal:

```bash
flashdata auth finish --wait --timeout 10m
flashdata auth status
```

The CLI saves and verifies the approved credentials. If approval or installation is interrupted, keep the same profile and run `auth finish --wait` again. A timeout ends local waiting; it does not create a new trial or discard saved progress.

## Use an existing API key

If you already have an API key, import it instead of starting a trial:

```bash
flashdata auth login
```

Paste the key at the hidden prompt. You can manage keys in the [FlashData console](https://flashdata.dev/app/api-keys). For automation, pipe a secret manager's output to `flashdata auth login --api-key-stdin --non-interactive`.

`FLASHDATA_API_KEY` overrides saved access for the current command. Query commands do not save it automatically. REST API keys and MCP OAuth tokens are not interchangeable. See [credential storage](security.md).

## Check access and recover

```bash
flashdata auth status
flashdata sources list
flashdata account balance
```

These commands show access, permitted sources and the current shared balance without submitting a paid query.

- If registration fails, rerun `init` with the same profile and unchanged input to recover the saved request.
- If access expires or credits run out, follow the recovery guidance from `auth status` or import a valid key. `init` does not replace expired or unusable credentials with another trial.
- Keep the same profile and state directory when recovering queries, downloads or browser approval. See [run recovery](runs.md).
- `auth logout` removes local credentials and retains run history. Revoke a remote key in your account if needed; local logout does not revoke it.
