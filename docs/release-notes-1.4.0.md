# FlashData CLI 1.4.0

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Use optional `session_token` credentials for S3, S3-compatible storage, OSS, COS,
BOS and TOS in download, batch and bulk storage configurations. Existing access
key configurations remain compatible; GCS authentication is unchanged.

- Pin the official Go SDK v1.3.0, including storage validation and token redaction.
- Keep session tokens out of CLI errors, diagnostics and rendered responses.
- Include an STS storage example and updated bundled agent guidance for expiring
  credentials and paused bulk delivery.

Credentials must stay valid until upload completes. The CLI passes the credentials
you supply and does not create or automatically refresh STS sessions.

Standalone archives cover Linux amd64/arm64, macOS Intel/Apple Silicon and Windows
x64. Python installer wheels and package-manager channels are prepared separately.
Existing profiles, saved runs and storage configurations remain compatible.
