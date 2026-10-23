# FlashData CLI v1.1.2

[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)

Default `flashdata auth login` now works on fresh headless Linux profiles when the
desktop keyring is unavailable, including servers with a configured D-Bus session.
The CLI automatically saves credentials in a private file and reuses it on later
commands. `flashdata init` uses the same storage selection.

- Private credential files have owner-only permissions (0600), in private directories (0700). They are not encrypted.
- Existing encrypted profiles keep using their original keyring. `--credential-store os` still requires OS protection.
- Existing credentials, run history and accepted-job recovery are preserved.
- JSON, plain-text and Markdown response options and the v1.1.1 environment-key fix remain available.

## Upgrade

Upgrade with your existing installer, or download an archive from the
[v1.1.2 release](https://github.com/flashdatadev/flashdata-cli/releases/tag/v1.1.2).
For uv tool installations:

```bash
uv tool upgrade flashdata-cli
flashdata --version
flashdata auth login
```

See the [installation guide](installation.md) for pip, uv pip, Homebrew, signed
APT/RPM and standalone downloads. Keep your profiles and run history. A profile
already saved with `--credential-store file` continues working without that flag.

This release uses Go SDK v1.1.0. It includes five platform archives, documentation,
a build report and SHA256SUMS. `flashdata --version` records the tagged source
commit. Previous tags and package versions remain available unchanged.

## 简体中文

1.1.2 修复无桌面 Linux 服务器上的默认登录：即使配置了 D-Bus 会话，桌面密钥环不可用时，
新配置也会自动使用私有文件保存凭据。后续命令无需额外参数；`flashdata init` 同样适用。
文件权限为 0600、目录权限为 0700，文件内容不加密。已有加密配置继续使用原密钥环；
显式选择 `--credential-store os` 时不会回退到文件存储。

请通过原有安装方式升级，保留凭据和执行历史。已经使用文件存储的配置可以继续直接运行
`flashdata auth login`、`flashdata auth status` 和查询命令。
