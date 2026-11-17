# Download storage setup

`youtube_download` uploads into storage you control. The `storage` object identifies a bucket or container and supplies the credentials FlashData needs for that upload.

## Amazon S3 example

Save this structure in a private local JSON file. Replace the placeholder values through your normal secret-management process. For CLI commands, pass its path with `--storage-config`; the optional Python example reads the path from `FLASHDATA_STORAGE_FILE`. Do not commit the filled file or paste its credentials into chat.

```json
{
  "storage_type": "s3",
  "storage_url": "your-media-bucket",
  "region": "us-east-1",
  "object_prefix": "youtube-clips",
  "access_key": "YOUR_ACCESS_KEY",
  "secret_key": "YOUR_SECRET_KEY"
}
```

For S3, `storage_url` is the bucket name, without `s3://` or a folder suffix. Put folders in `object_prefix`, using path segments such as `media/youtube`. The destination and credentials must allow writing the requested objects.

## Temporary STS credentials

For S3, S3-compatible providers that issue temporary credentials, OSS, COS, BOS and
TOS, add `session_token` alongside the matching temporary `access_key` and
`secret_key`. Copy the [STS storage example](../assets/storage-sts.json) into the
user's private storage configuration and replace its placeholders through their
secret-management process. AWS/BOS/TOS call the token SessionToken; OSS calls it
SecurityToken; COS calls it Token and pairs it with TmpSecretId/TmpSecretKey.

Keep the provider's usual bucket, region and endpoint fields. Azure uses `sas_token`
and GCS uses `service_account`; do not add `session_token` to those configurations.
Omit it when using long-term access keys. Use an updated CLI/SDK/MCP release whose
storage schema accepts `session_token`.

The same configuration applies to single and batch downloads, bulk download
`options.storage`, and bulk data `options.result_storage`. Credentials must remain
valid through queueing, processing, upload and retries. FlashData does not refresh
temporary credentials automatically. For a paused data bulk job whose input scan
has finished, update the result-storage key pair and token together, then resume
that existing job. Keep tokens out of chat, logs and recovery records.

## Other providers

| `storage_type` | Configuration |
| --- | --- |
| `s3_compatible` | Bucket name, `access_key`, `secret_key`, provider `endpoint_url` and region as required. |
| `gcs` | Bucket name and `service_account` with `project_id`, `client_email`, and PEM `private_key`. |
| `azure` | Container name in `storage_url`, `account_name`, and a container `sas_token` with write permission and expiry. |
| `oss` | Bucket name, `access_key`, `secret_key` and provider endpoint/region. |
| `cos` | Bucket name including APPID, `access_key`, `secret_key`, and region such as `ap-guangzhou`. |
| `bos` | Bucket name, `access_key`, `secret_key`, and supported region (`bj`, `su`, `hkg`). |
| `tos` | Bucket name, `access_key`, `secret_key`, and provider region. |

Use the current [download documentation](https://flashdata.dev/docs/sources/youtube_download) and the provider's settings for the exact account configuration. Credentials are sent to FlashData for upload and may appear in MCP tool history; keep access limited to the intended destination. The returned object follows your bucket's access and retention settings. Uploading does not make the object public.
