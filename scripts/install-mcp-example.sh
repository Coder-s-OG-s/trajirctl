#!/usr/bin/env bash
# Print a ready-to-paste mcpServers block after you build trajir-mcp.
# Usage:
#   ./scripts/install-mcp-example.sh /abs/path/to/trajir-mcp /abs/path/to/project
set -euo pipefail

bin="${1:-}"
root="${2:-}"
if [[ -z "$bin" || -z "$root" ]]; then
  echo "usage: $0 /abs/path/to/trajir-mcp /abs/path/to/project" >&2
  exit 2
fi
if [[ "$bin" != /* || "$root" != /* ]]; then
  echo "error: both paths must be absolute" >&2
  exit 2
fi

python3 - "$bin" "$root" <<'PY'
import json, sys
bin_path, root = sys.argv[1], sys.argv[2]
print(json.dumps({
    "mcpServers": {
        "trajectory-ir": {
            "command": bin_path,
            "args": [],
            "env": {"TRAJIR_MCP_ROOT": root},
        }
    }
}, indent=2))
PY
