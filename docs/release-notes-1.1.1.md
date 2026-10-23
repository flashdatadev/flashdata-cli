# FlashData CLI v1.1.1

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Fix consecutive commands using `FLASHDATA_API_KEY` on a new profile. After the
first query, later commands no longer fail with `credentials_missing` while the
environment key is still set.

- Save profile origins and run history without saving the environment key.
- Cover realtime queries, async and batch submissions, and bulk workflows.
- Preserve existing saved credentials, origin checks and cached-run reuse.
- Keep JSON, plain-text and Markdown response options unchanged.

## Upgrade

Upgrade through your existing installer, or download the archive for your platform
from the [v1.1.1 release](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.1.1).
Keep your profiles and run history. See the [installation guide](installation.md)
for pip, uv, Homebrew, signed APT/RPM and standalone downloads.

If a profile created by v1.1.0 already reports `credentials_missing`, run
`flashdata auth login` to verify and save your key. Supply the original origin
flags when using custom service origins. Your existing run IDs and results remain
available; do not delete state or resubmit accepted work. Scripts can import a key
through `auth login --api-key-stdin`.

This patch uses Go SDK v1.1.0. The release includes five platform archives, matching
documentation, a build report and `SHA256SUMS`. `flashdata --version` identifies
the tagged source commit. Earlier tags and package versions remain unchanged.

## 简体中文

1.1.1 修复新配置使用 `FLASHDATA_API_KEY` 时，首次查询成功后，后续命令仍报
`credentials_missing` 的问题。配置会保存服务地址和执行历史，但不会保存环境变量中的密钥。
实时查询、异步和批量提交、批量导出均适用；JSON、纯文本和 Markdown 选项保持不变。

请通过原有安装方式升级并保留配置和执行历史。如果 1.1.0 创建的配置已受影响，运行
`flashdata auth login` 验证并保存密钥；使用自定义服务地址时，传入原有地址参数。
已有运行 ID 和结果会保留，无需删除配置或重复提交任务。
