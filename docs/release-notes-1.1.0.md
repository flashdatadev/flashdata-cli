# FlashData CLI v1.1.0

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Choose JSON, plain text or Markdown with independent `--accept` and
`--respond-with` flags. JSON remains the default.

- Search, YouTube and generic query commands support all four content formats.
- Async/batch receipts and job status support `--accept text/plain`.
- Use `jobs result` or async/batch `--wait` for Markdown completed results.
- Plain text goes directly to stdout; add `--json` for the CLI envelope or
  `--output FILE` to save the selected body.
- Saved runs remember format choices. Reformat completed background results
  without submitting another query. A failed save preserves the previous cache.
- Redact credentials escaped by Markdown, including older cached responses and
  file exports, while retaining source links.
- Pin Go SDK v1.1.0 and preserve accepted IDs, uncertain receipts, access checks
  and durable recovery.

## Install or upgrade

Download the archive for your platform from the
[official release](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.1.0).
Standalone binaries need no Go, Python or Node.js. Keep your existing profiles
and state directory when replacing the executable.

The [installation guide](installation.md) covers Homebrew, signed APT/RPM,
manual archives and Windows. Go 1.25+ users can install the source tag:

```bash
go install github.com/flashdatadev/flashdata-cli/cmd/flashdata@v1.1.0
```

The release includes Linux amd64/arm64, macOS Intel/Apple Silicon and Windows x64
archives, matching English/Chinese documentation, a build report and `SHA256SUMS`.
`flashdata --version` identifies the exact tagged source commit. Existing release
assets remain unchanged. Bulk exports and media keep their existing formats.

## 简体中文

1.1.0 新增独立的 `--accept` 和 `--respond-with` 参数，默认仍为 JSON。
实时内容和已完成结果支持 Markdown；后台提交、批量提交及任务状态支持纯文本。
执行历史保留格式设置，切换已完成后台结果的格式不会重复提交；保存失败时保留旧缓存。
修复 Markdown 转义后的凭据脱敏，涵盖旧缓存读取和导出。五个平台均提供独立可执行文件。
