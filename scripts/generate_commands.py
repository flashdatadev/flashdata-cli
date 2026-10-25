"""Generate the command reference from an offline local CLI executable."""

import argparse
import json
import os
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def cell(value):
    return str(value).replace("|", "\\|").replace("\n", " ")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=Path)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    env = os.environ.copy()
    env.pop("FLASHDATA_API_KEY", None)
    output = subprocess.check_output(
        [str(args.binary.resolve()), "commands", "--json", "--non-interactive"],
        env=env,
        encoding="utf-8",
        timeout=15,
    )
    manifest = json.loads(output)
    if not manifest.get("ok") or manifest.get("schemaVersion") != 1:
        parser.error("expected a version-1 command manifest")
    lines = [
        "# Command reference",
        "",
        "[FlashData](https://flashdata.dev) · [API documentation](https://flashdata.dev/docs)",
        "",
        "Generated from `flashdata commands --json`. Every operational command supports the global profile, state directory, credential protection, timeout and output-mode flags. Keys are accepted only through hidden input/stdin or an explicitly configured environment variable.",
        "",
        "Start with `flashdata setup` for an agent, or `flashdata init --name NAME` for terminal-only use. A new profile without credentials can request temporary access before signup, subject to trial availability and eligibility; existing access is reused. You do not need both commands. See [onboarding](onboarding.md) for trial limits and connecting an account, and [agent setup](agent-setup.md) for offline installation with `--skip-auth`.",
        "",
        "See [response formats](response-formats.md) for JSON, plain-text and Markdown options. `discover plans` and `discover status` read public endpoints without an API key.",
        "",
    ]

    def visit(command, parent=""):
        name = f"{parent} {command['name']}".strip()
        lines.extend(
            [
                f"## {name}",
                "",
                command["description"],
                "",
                f"Usage: `{command['use']}`",
                "",
            ]
        )
        if command["flags"]:
            lines.extend(
                [
                    "| Flag | Type | Default | Purpose |",
                    "| --- | --- | --- | --- |",
                ]
            )
            for flag in command["flags"]:
                lines.append(
                    f"| `--{cell(flag['name'])}` | {cell(flag['type'])} | `{cell(flag['default'])}` | {cell(flag['description'])} |"
                )
            lines.append("")
        for child in command["commands"]:
            visit(child, name)

    visit(manifest["data"])
    text = "\n".join(lines).rstrip() + "\n"
    destination = ROOT / "docs/commands.md"
    if args.check:
        if destination.read_text(encoding="utf-8") != text:
            parser.error("command reference is stale; run again without --check")
        print("Command reference matches the executable.")
    else:
        destination.write_text(text, encoding="utf-8", newline="\n")
        print("Updated docs/commands.md from the offline command tree.")


if __name__ == "__main__":
    main()
