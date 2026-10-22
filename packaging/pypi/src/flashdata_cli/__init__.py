"""Launch the bundled FlashData executable without changing CLI behavior."""
import os
import subprocess
import sys
from pathlib import Path


def main():
    """Pass arguments, streams, environment and exit status to the native CLI."""
    binary = Path(__file__).resolve().parent / "bin" / ("flashdata.exe" if os.name == "nt" else "flashdata")
    arguments = [str(binary), *sys.argv[1:]]
    try:
        if os.name != "nt":
            # Keep PID, terminal and signal handling under the native executable.
            os.execv(str(binary), arguments)
        # Windows has no Unix process replacement. The child shares this console
        # and inherits stdin/stdout/stderr; Ctrl+C reaches it directly.
        child = subprocess.Popen(arguments)
        while True:
            try:
                return child.wait()
            except KeyboardInterrupt:
                # Let the CLI finish its own interruption/recovery handling.
                continue
    except OSError as error:
        # Do not echo arguments or the environment, which can contain secrets.
        print("Cannot start bundled FlashData executable: " + (error.strerror or "operating system error"), file=sys.stderr)
        return 127 if isinstance(error, FileNotFoundError) else 126


if __name__ == "__main__":
    raise SystemExit(main())
