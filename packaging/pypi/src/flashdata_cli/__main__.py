"""Allow python -m flashdata_cli as well as the installed flashdata command."""
from . import main

raise SystemExit(main())
