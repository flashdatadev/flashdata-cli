# Runnable inputs

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Use query.json with query --input, query-flat.json for the spec-compatible flat form, youtube-metadata.json for async recovery, batch.json for batch submit, and ids.txt for bulk data jobs. These files contain public query/video examples, not credentials.

bulk-create.json and bulk-update.json are runnable metadata inputs for manual creation and a pre-upload options patch. Feed them through --input - using Bash redirection or PowerShell Get-Content -Raw; replace the remote bulk ID in later commands. Metadata patches after a scan has completed are restricted by the backend; use only the credential-only recovery patch documented for an eligible paused customer-delivery job.

bulk-options.json and download-options.json show option shapes. Copy a chosen option into a private file before passing --options. Copy storage.example.json to a private location outside the checkout and replace every placeholder with your actual authorized destination configuration. The CLI requires private storage/config permissions; Unix example files are intentionally not treated as a secret vault.

Offline command and compiled-binary tests cover the onboarding, query, batch, bulk and recovery workflows, using generated private fixture storage instead of real cloud credentials. Tests also validate example JSON and check that both README translations use identical command blocks. Secret-manager names, private paths, remote IDs and storage values in the docs are placeholders to replace. Connecting these examples to production incurs ordinary API charges; a dry-run validates and reads metadata without a paid submission.
