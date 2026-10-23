# Credential and storage model

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

The expanded Linux automatic fallback below is unreleased. With v1.1.1, use `flashdata --credential-store file auth login` if the desktop keyring is unavailable; later commands reuse the saved storage automatically.

| Platform | Automatic OS protection | Headless fallback |
| --- | --- | --- |
| Windows amd64 | Native current-user DPAPI, private owner/SYSTEM DACLs | DPAPI requires the same user profile; explicit file mode is available |
| macOS amd64/arm64 | Native Security.framework Keychain key with UI disabled, AES-256-GCM vault | Use an unlocked accessible keychain or explicitly choose private file mode |
| Linux amd64/arm64 | Unlocked Secret Service over session D-Bus, AES-256-GCM vault; never calls Unlock/Prompt | New profiles automatically use private files when the session bus or keyring is unavailable, even if a D-Bus address is set |

OS failures never silently downgrade an existing store. macOS uses pure-Go native bindings without CGo or a `security` subprocess; Windows uses native DPAPI calls without PowerShell. Secrets do not enter subprocess arguments or shell commands. Linux keyring calls do not ask for GUI unlock. The master key is distinct from the API key and bound to the canonical state-root/profile identity; copying an OS-protected vault to another user/location may require explicit migration/key import.

On a fresh Linux profile, `auth login` and `init` work without storage flags on headless servers. Automatic mode tries the configured keyring and uses a private file if it cannot save there. The selected storage is reported as `credentialProtection` and reused by later commands. `--credential-store os` requires the keyring and disables this fallback. Existing keyring profiles, missing vaults with recovery history, and canceled operations do not fall back to file storage.

`--credential-store file` is a documented protected plaintext mode: JSON is base64-encoded inside a versioned envelope, not encrypted. Use persistent storage owned by the scheduled account. Unix requires private directories (0700), private regular files (0600) and current-user ownership; symlink state paths are refused. Windows checks ACLs/ownership and writes private output ACLs. State writes use a synced temporary file, atomic replacement and directory synchronization where supported. Existing unsafe/corrupt state is never reset.

Defaults live outside repositories: Windows/macOS use the OS user configuration directory under FlashData; Linux uses an absolute XDG_STATE_HOME/flashdata or ~/.local/state/flashdata. Each lowercase profile has an independent vault, lock, run journal and results directory. `--state-dir` is an explicit persistent-location override; keep it out of project repositories. Do not delete state to repair authentication or uncertainty.

The same OS user, administrators/root, a compromised process or a malicious custom SDK transport are outside this local secrecy boundary. OS encryption protects copied credential blobs; it is not an isolation boundary against the logged-in user. Successful provider data may itself be sensitive. Store exported results appropriately.

API keys, registration/setup/handoff proofs, cloud access/secret/private keys and signed capabilities are redacted from CLI diagnostics and sanitized metadata. Provider URLs and large numbers are otherwise retained. Successful data is never evaluated as shell commands, prompts or code. SDK requests pin service origins and authenticated signed-storage origins, isolate keys/cookies from storage, and never retry paid submissions automatically.

Customer-storage config and option files must be private and outside source control. Unix: `chmod 600 /private/flashdata/storage.json`. On Windows use a directory accessible only to the scheduled user/SYSTEM, or set the file DACL explicitly using your organization's approved ACL tooling. The example storage file contains only placeholders and must not be used unchanged.

Generic query/batch JSON files containing credential fields or callback URLs require the same private permissions; stdin is supported when a secret manager supplies the payload. Callback destinations are omitted from sanitized metadata because webhook URLs can contain secrets. Public credential-property schema definitions remain intact; they contain no secret values. Discovery also describes strict top-level fields and the optional background callback field enforced by SDK runtime validation. Windows exported files receive their private security descriptor at creation time, before any result bytes are written.

An environment override changes only the current command's access. Keep `FLASHDATA_API_KEY` set for each command that needs authentication. A new environment-only profile saves its origins and run history without saving the key, so consecutive commands work without `auth login`. Updates preserve any previously saved key, proofs and phase. Re-importing the same saved key is a verified reuse operation. Linux accepts an already configured local Unix D-Bus endpoint, never auto-launches a session bus, and refuses remote/autolaunch transports for Secret Service.

CLI 1.1.0 could report `credentials_missing` after the first environment-only query. If your profile is already affected, keep its state and run `flashdata auth login` to verify and save your key. For custom service origins, supply the same origin flags used for the original query. Existing run IDs and results remain available; do not delete the profile or resubmit accepted work.

An explicit auth login into an environment-only profile persists the chosen verified key while retaining accepted-job recovery capabilities; it does not require a destructive replacement. Existing run/account bindings still control which records that key can use.

`auth logout` clears API keys, proofs and saved upload capabilities while retaining history. It reports that remote access is not revoked and whether an environment key remains set. Remove environment variables separately and revoke remote keys in your FlashData account when desired. `auth login --replace` explicitly replaces prior onboarding/upload capabilities; prefer a new profile if old work still needs those capabilities.
