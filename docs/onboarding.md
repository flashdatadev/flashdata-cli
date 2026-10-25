# Onboarding and recovery

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

## Try before signing up

You can start with temporary trial access without a FlashData account or an API key of your own. For a new profile without configured credentials, the CLI requests trial access and saves it locally. Trials are subject to availability and eligibility; existing access is reused.

For an AI agent, install the CLI and run setup:

```sh
uv tool install flashdata-cli
flashdata setup
```

Setup requires CLI 1.2.0 or newer. While the PyPI update is pending, use a [native installer or release download](installation.md#homebrew-linux-and-macos).

Choose Codex, Claude Code or Cursor and confirm the directories. `setup` installs skills and configures access in one step. You do not also need to run `init`.

For terminal-only use, run this **instead of setup**:

```bash
flashdata init --name "my-cli"
```

After either command, try your first search:

```bash
flashdata search "battery recycling startups" --num 5 --accept text/plain --json=false
```

Setup and init make no paid queries. Searches use trial Credits. The default trial sources are Google Search (`google_search`) and YouTube metadata (`youtube_metadata`). Installing a skill does not grant access to its data source.

Before your first `setup` or `init`, you can add `--scopes youtube_search,youtube_metadata,youtube_transcript` to request different trial sources. The service decides which sources are approved. Passing `--scopes` when reusing existing access does not expand its permissions.

`setup --skip-auth` only installs and checks skills offline. It does not create a trial, check access or change credentials. Queries still need configured access; rerun setup without `--skip-auth` when ready. See [agent setup](agent-setup.md) for more options.

## Check trial limits

```bash
flashdata discover --json
flashdata auth status
flashdata sources list
flashdata account balance
```

These commands show current trial availability and limits, your access and expiry, permitted sources and remaining Credits without submitting a paid query. Trials have a limited shared credit allowance. Per-source request counts are alternative uses of that same allowance, not separate quotas.

If trial registration is unavailable or you are ineligible, sign up and import an API key with `flashdata auth login`. If temporary access expires or Credits run out, check `auth status` and connect your account below. Rerunning setup or init does not create a replacement trial.

## Connect your FlashData account

You can sign up later, when you are ready to continue beyond the trial. Keep the same CLI profile and start the account connection:

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

If this profile already has trial access or a saved registration attempt, use `flashdata auth login --replace` to replace it with your key. To connect the trial through browser approval instead, use `auth claim` and `auth finish` as described above.

On a fresh headless Linux setup, login automatically uses a private credential file if the desktop keyring is unavailable. Later commands reuse it without extra flags. The file has owner-only permissions and is not encrypted; see [credential storage](security.md).

`FLASHDATA_API_KEY` overrides saved access for the current command. Query commands do not save it automatically. REST API keys and MCP OAuth tokens are not interchangeable. See [credential storage](security.md).

## Recover access

- If registration fails, rerun `init` with the same profile and unchanged input to recover the saved request.
- If access expires or credits run out, follow the recovery guidance from `auth status` or import a valid key. `init` does not replace expired or unusable credentials with another trial.
- Keep the same profile and state directory when recovering queries, downloads or browser approval. See [run recovery](runs.md).
- `auth logout` removes local credentials and retains run history. Revoke a remote key in your account if needed; local logout does not revoke it.
