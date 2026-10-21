# Onboarding and recovery

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Use `auth login` with hidden terminal input or `--api-key-stdin` for an existing account. Do not supply keys in arguments. An environment key overrides saved access for the current command and is never implicitly persisted by query/bulk operations. It does not replace stored candidate credentials during `auth finish`.

Saving recovery metadata under an environment override preserves the original profile's credentials and onboarding phase. Re-importing the same saved key verifies and reuses the profile without removing claim proofs or upload capabilities. auth status also reports registering/logged_out/environment-required local phases without attempting registration; a pending registration points back to init with its saved proof. Scope-list whitespace is normalized consistently on an explicit registration retry.

`init --name NAME` explicitly starts onboarding only for a genuinely new profile. It discovers protocol 2 and availability, generates a 256-bit proof, atomically saves the proof and exact registration request, then calls the SDK. Recovery reuses the saved request/proof without requiring flags again. Changing a pending request conflicts. A marker, corrupted/missing vault, logged-out profile or failed/expired existing key prevents a new automatic trial.

Registration returns ordinary API-key credentials. Setup proof, registration proof and handoff proof are scoped to their control operations; they are never sent as Data API keys. Client labels and plan names convey no authorization. The public trial grant/request-equivalents are distinct from current account balance, and no retry refills a grant.

`auth claim` stores the returned claim privately, constructs its URL from the pinned web origin and validated fixed claim path, and displays the confirmation code separately. No proof/key/code goes into the browser URL. The browser owns sign-in, code confirmation, scope review and funding approval; the CLI cannot impersonate a browser session. The CLI reports the URL rather than automatically opening a browser.

`auth finish --wait --timeout 10m` polls only recoverable statuses, honoring the server interval. After delivery it commits the candidate key, generation and original handoff proof in the protected vault. Its fresh child is the same absolute executable, supplied only a state directory/profile and a random verification challenge. The child does not inherit an environment API key, takes no API key in arguments, reads the vault without claiming the parent's writer lock, and verifies connection ID, generation, valid lifecycle, ready access and owned state. ACK follows only a matching challenge response. Installation is read back from the server before obsolete secrets are removed.

A failed child sends no ACK. A lost ACK response retains the candidate/proof for explicit recovery. An expired or confirmed handoff without a locally installed key needs account reconnect/recovery, never a new anonymous trial. Read `auth status`; it can still display pending onboarding state when an old temporary key has been fenced/revoked during claiming.

Imported-key ownership uses GET `/v2/api-access/identity`. A 404 permits ordinary one-off API-key use with ownership unknown, but blocks recurring mode. A known claimed connection can instead verify its current key through its authoritative connection identity endpoint; recurring requires owned, valid, ready and installed. Failures other than a missing endpoint are not guessed away. Revalidation happens before each execution and before cached result reads.

The new general identity endpoint is a local backend patch awaiting deployment; onboarding is feature-gated. This project does not configure production flags, publish or deploy anything.
