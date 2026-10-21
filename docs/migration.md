# Explicit legacy migration

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

The prior Node script writes a version-1 protection envelope around client state version 2 in connection.json. `auth migrate --from PRIVATE_CONNECTION_JSON --profile migrated` reads only this explicit format, using native DPAPI on Windows or the private-file legacy format on Unix. The source is unchanged. Unsupported versions, unsafe permissions and corrupt envelopes fail closed.

Migration validates/pins origins, preserves registration input/proof and pending claim/handoff phases, imports completed results and retains uncertain-operation metadata. It writes a legacy replay barrier before publishing the migrated credentials. A partial migration with a missing vault still blocks init from creating replacement access; rerun the explicit migration to continue.

A private migration marker pins an interrupted import to the original source digest before the replay barrier is written. Another source cannot replace that import or a profile with unrelated history and a missing vault. Repeating an already completed migration is a no-op that retains the current credentials, phase and history, including subsequent logout or claiming; obsolete source credentials are never restored as a retry.

The script retained original run IDs only as SHA-256 keys. The CLI cannot reconstruct those IDs or prove equivalence with a newly normalized request, so it imports readable legacy-HASH records and keeps the original hash tombstones. A later command using one of those original IDs refuses automatic submission. Cached legacy data requires current access/account verification; if ownership could not be established during migration, inspect metadata and recover through your account rather than inventing an association.

This command does not convert a corrupt vault, switch an existing OS-protection mode, reset history, obtain a trial or revoke remote access. To move to another protection mode, explicitly import an existing key into a fresh profile and keep the old profile for recovery/history; pending onboarding needs its original store until installation completes.
