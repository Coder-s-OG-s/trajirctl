# trajirctl (+ Trajectory IR MCP host packaging)

Human CLI for local [Trajectory IR](https://github.com/Coder-s-OG-s/Trajectory-IR)
workdirs, plus host-config packaging for the **`trajir-mcp`** agent server.

| Surface | Binary / config | Audience |
|---------|-----------------|----------|
| MCP (agents) | `trajir-mcp` from Trajectory-IR + configs in [`integrations/mcp/`](integrations/mcp/) | Claude Code, Cursor, … |
| CLI (humans / scripts) | `trajirctl` (this module) | Terminal operators |

Both wrap the same Go SDK (`Trajectory-IR/go`). MCP enforces
`TRAJIR_MCP_ROOT` workspace confinement; `trajirctl` does not — see
`docs/superpowers/specs/2026-08-27-trajirctl-design.md` in Trajectory-IR.

## Install CLI

```bash
go install github.com/Coder-s-OG-s/trajirctl@v0.1.0
# or @latest
```

Requires Go 1.25+.

## Wire MCP for agents (`TRAJIR_MCP_ROOT`)

1. Build the server from Trajectory-IR:

```bash
git clone https://github.com/Coder-s-OG-s/Trajectory-IR.git
cd Trajectory-IR/go
go build -o trajir-mcp ./cmd/trajir-mcp
```

2. Point your host at the binary with an absolute project root:

```json
{
  "mcpServers": {
    "trajectory-ir": {
      "command": "/absolute/path/to/trajir-mcp",
      "args": [],
      "env": {
        "TRAJIR_MCP_ROOT": "/absolute/path/to/project"
      }
    }
  }
}
```

Templates and details: [`integrations/mcp/README.md`](integrations/mcp/README.md).

Helper to print a filled config:

```bash
# Unix
./scripts/install-mcp-example.sh /abs/path/to/trajir-mcp /abs/path/to/project

# Windows PowerShell
.\scripts\install-mcp-example.ps1 -Bin C:\abs\path\to\trajir-mcp.exe -Root C:\abs\path\to\project
```

### MCP tools

| Tool | Purpose |
|------|---------|
| `trajectory_status` | Counts / seals / paths for a workdir trajectory |
| `trajectory_export_tir` | Export thin or fat `.tir` |
| `trajectory_import_tir` | Load + hash-verify (does **not** write NodeLog) |
| `trajectory_verify_signature` | Signature check; unsigned OK unless `require_signature` |

## Day-to-day CLI usage (`trajirctl`)

Inspect the same local workdirs agents use:

```bash
export TRAJIR_WORKDIR=./my-data
export TRAJIR_TENANT=acme
export TRAJIR_TRAJECTORY=trip-1

trajirctl status
trajirctl nodes list
trajirctl nodes show --id <NODE_ID>
trajirctl export --dest ./out.tir --mode thin
trajirctl import --path ./out.tir
trajirctl verify --path ./out.tir
trajirctl verify --path ./out.tir --require-signature   # exits 1 if unsigned/invalid
```

### Commands

```
trajirctl status --workdir DIR --tenant ID --trajectory ID [--json]
trajirctl export --workdir DIR --tenant ID --trajectory ID --dest PATH [--mode thin|fat] [--json]
trajirctl import --path PATH [--src PATH] [--json]
trajirctl verify --path PATH [--src PATH] [--require-signature] [--json]
trajirctl nodes list --workdir DIR --tenant ID --trajectory ID [--json]
trajirctl nodes show --workdir DIR --tenant ID --trajectory ID --id NODE_ID [--json]
```

`--workdir` / `--tenant` / `--trajectory` fall back to `TRAJIR_WORKDIR` /
`TRAJIR_TENANT` / `TRAJIR_TRAJECTORY`. `--workdir` defaults to `.`.
`--path` is preferred for packages; `--src` is an accepted alias (MCP uses `path`).

Exit codes: `0` success; `1` error (including verify `failed` and nodes show
not-found); `2` usage.

## Test

```bash
go test ./...
go vet ./...
go build -o trajirctl .
```

## Scope

v1 is local-workdir only — no remote/deployed backend. License: Apache-2.0.
