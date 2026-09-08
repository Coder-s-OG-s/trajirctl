# Trajectory IR MCP host packaging

This folder is the **host-facing** packaging surface for Trajectory IR’s MCP
server. IR semantics and the `trajir-mcp` binary live in
[Trajectory-IR](https://github.com/Coder-s-OG-s/Trajectory-IR)
(`go/cmd/trajir-mcp`, `go/trajir/mcp`). Here we only document how to wire the
server into agent hosts and how operators use `trajirctl` beside it.

```text
AI host (Claude Code, Cursor, …)
        │  MCP stdio
        ▼
  trajir-mcp   (Trajectory-IR)   ← set TRAJIR_MCP_ROOT
        │
        ▼
  local workdir (nodes.sqlite / memo.sqlite)

Human / scripts
        │
        ▼
  trajirctl    (this repo)       ← no MCP root jail
```

## Build the MCP server

```bash
git clone https://github.com/Coder-s-OG-s/Trajectory-IR.git
cd Trajectory-IR/go
go build -o trajir-mcp ./cmd/trajir-mcp
```

Requires Go 1.25+.

## Wire `TRAJIR_MCP_ROOT` (required for safe agent use)

`TRAJIR_MCP_ROOT` is the approved workspace root. Every MCP `work_dir`,
`dest`, and `path` argument must resolve under it (CWE-73 defense against
prompt-injected path steering). When unset, the process cwd is used as the
root — always set it explicitly in host configs.

| Variable | Meaning |
|----------|---------|
| `TRAJIR_MCP_ROOT` | Absolute path to the project/workspace the agent may touch |

Example host configs in this directory:

| File | Host |
|------|------|
| [`claude-code.mcp.json`](./claude-code.mcp.json) | Claude Code / Cursor-style `mcpServers` |
| [`cursor.mcp.json`](./cursor.mcp.json) | Cursor MCP config snippet |

Replace `__TRAJIR_MCP_BIN__` and `__PROJECT_ROOT__` with absolute paths before
installing.

### Claude Code / Cursor style

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

## MCP tools (from Trajectory-IR)

| Tool | Purpose |
|------|---------|
| `trajectory_status` | Node counts by kind, seal count, paths |
| `trajectory_export_tir` | Export thin (default) or fat `.tir` |
| `trajectory_import_tir` | Load + hash-verify `.tir` (no NodeLog write) |
| `trajectory_verify_signature` | Optional signature verify; unsigned OK unless `require_signature` |

Common args: `work_dir`, `tenant_id`, `trajectory_id`, `path` / `dest`.

## Day-to-day: agents vs humans

| Task | Use |
|------|-----|
| Agent reads/exports/verifies inside a project | **MCP** (`trajir-mcp` + `TRAJIR_MCP_ROOT`) |
| You inspect a workdir at the terminal | **`trajirctl`** (this binary) |
| Script CI checks on a `.tir` | **`trajirctl verify --require-signature`** |

`trajirctl` deliberately does **not** enforce `TRAJIR_MCP_ROOT`: a person
typing `--workdir ./data` is the trust boundary. See the design spec in
Trajectory-IR: `docs/superpowers/specs/2026-08-27-trajirctl-design.md`.

## Smoke-test the MCP binary

```bash
# stdio server waits for an MCP host — use your host’s MCP inspector,
# or confirm the binary starts and exits cleanly on EOF:
printf '' | ./trajir-mcp ; echo exit:$?
```

Full contract: [Trajectory-IR `docs/INTEGRATIONS.md`](https://github.com/Coder-s-OG-s/Trajectory-IR/blob/main/docs/INTEGRATIONS.md).
