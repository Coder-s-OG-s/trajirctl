# trajirctl

Human operated CLI for local [Trajectory IR](https://github.com/Coder-s-OG-s/Trajectory-IR)
workdirs, plus **MCP host packaging** for the agent facing `trajir-mcp` server.

| Surface | What you use | Audience |
|---------|--------------|----------|
| **MCP** (agents) | `trajir-mcp` from Trajectory-IR + configs in [`integrations/mcp/`](integrations/mcp/) | Claude Code, Cursor, and other MCP hosts |
| **CLI** (humans / scripts) | `trajirctl` (this repository) | Terminal operators and CI |

Both call the same Go SDK (`github.com/Coder-s-OG-s/Trajectory-IR/go`). The MCP
server enforces `TRAJIR_MCP_ROOT` workspace confinement (CWE-73). `trajirctl`
does **not**: a person typing `--workdir` is the trust boundary. Design
rationale:
[`docs/superpowers/specs/2026-08-27-trajirctl-design.md`](https://github.com/Coder-s-OG-s/Trajectory-IR/blob/main/docs/superpowers/specs/2026-08-27-trajirctl-design.md).

**License:** [Apache 2.0](LICENSE) · **Go:** 1.25+ · **Latest tag:** [`v0.1.0`](https://github.com/Coder-s-OG-s/trajirctl/releases/tag/v0.1.0)

---

## What is Trajectory IR?

Trajectory IR is a portable intermediate representation for agent runs: an
append only sequence of typed nodes. Before world changing tools execute, the
model’s plan for that step is **sealed**. Workdirs store that history in
`nodes.sqlite` (+ `memo.sqlite`); packages export as `.tir` for audit, handoff,
or verification outside the runtime that produced them.

This repo does **not** reimplement IR semantics. It wraps the public SDK the
same way `trajir-mcp` does, for humans instead of agents.

```text
AI host (Claude Code, Cursor, …)
        │  MCP stdio
        ▼
  trajir-mcp          ← Trajectory-IR repo; set TRAJIR_MCP_ROOT
        │
        ▼
  trajir/client + tir + NodeLog (SQLite workdir)

You / CI scripts
        │
        ▼
  trajirctl           ← this repo; no MCP root jail
```

---

## Prerequisites

- **Go 1.25+** (see `go.mod`)
- A local Trajectory IR **workdir** containing (or that will create)
  `nodes.sqlite` / `memo.sqlite`, plus known `tenant_id` and `trajectory_id`
- For MCP: a built `trajir-mcp` binary from
  [Trajectory-IR](https://github.com/Coder-s-OG-s/Trajectory-IR)

---

## Install `trajirctl`

```bash
go install github.com/Coder-s-OG-s/trajirctl@v0.1.0
# or track main:
go install github.com/Coder-s-OG-s/trajirctl@latest
```

From a clone of this repo:

```bash
git clone https://github.com/Coder-s-OG-s/trajirctl.git
cd trajirctl
go build -o trajirctl .
# Windows: go build -o trajirctl.exe .
```

Confirm:

```bash
trajirctl --help
```

---

## Quick start (CLI)

Point at a workdir that already has trajectory data (created by an agent, SDK
demo, or MCP tools):

```bash
export TRAJIR_WORKDIR=/absolute/path/to/workdir
export TRAJIR_TENANT=acme
export TRAJIR_TRAJECTORY=trip-1

trajirctl status
trajirctl nodes list --json
trajirctl export --dest ./out.tir --mode thin
trajirctl import --path ./out.tir
trajirctl verify --path ./out.tir
```

On Windows PowerShell:

```powershell
$env:TRAJIR_WORKDIR = "C:\path\to\workdir"
$env:TRAJIR_TENANT = "acme"
$env:TRAJIR_TRAJECTORY = "trip-1"

.\trajirctl.exe status
.\trajirctl.exe nodes list --json
```

If you do not have a workdir yet, create one with the Trajectory-IR Go client
(see [Go quickstart](https://github.com/Coder-s-OG-s/Trajectory-IR/blob/main/go/QUICKSTART.md))
or run a demo such as `go run ./examples/adoption_host` from `Trajectory-IR/go`.

---

## Wire MCP for agents (`TRAJIR_MCP_ROOT`)

Agents should talk to Trajectory IR through **MCP**, not by shelling out to
`trajirctl`.

### 1. Build `trajir-mcp`

```bash
git clone https://github.com/Coder-s-OG-s/Trajectory-IR.git
cd Trajectory-IR/go
go build -o trajir-mcp ./cmd/trajir-mcp
```

### 2. Configure the host with an absolute project root

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

| Variable | Meaning |
|----------|---------|
| `TRAJIR_MCP_ROOT` | Approved workspace root. Every MCP `work_dir`, `dest`, and `path` must resolve under this directory. When unset, process cwd is used: **always set it explicitly**. |

Templates:

| File | Use |
|------|-----|
| [`integrations/mcp/claude-code.mcp.json`](integrations/mcp/claude-code.mcp.json) | Claude Code / Cursor style `mcpServers` |
| [`integrations/mcp/cursor.mcp.json`](integrations/mcp/cursor.mcp.json) | Cursor MCP snippet |
| [`integrations/mcp/README.md`](integrations/mcp/README.md) | Full MCP packaging guide |

Generate a filled JSON blob:

```bash
# Unix
./scripts/install-mcp-example.sh /abs/path/to/trajir-mcp /abs/path/to/project

# Windows PowerShell
.\scripts\install-mcp-example.ps1 -Bin C:\abs\path\to\trajir-mcp.exe -Root C:\abs\path\to\project
```

### MCP tools

| Tool | Purpose |
|------|---------|
| `trajectory_status` | Node counts by kind, seal count, paths |
| `trajectory_export_tir` | Export thin (default) or fat `.tir` |
| `trajectory_import_tir` | Load and hash verify a `.tir` (**does not** write into NodeLog) |
| `trajectory_verify_signature` | Optional `trajir-pkg-sig-v1` verify; unsigned OK unless `require_signature` |

Common arguments: `work_dir`, `tenant_id`, `trajectory_id`, `path` / `dest`.

Contract reference:
[Trajectory-IR `docs/INTEGRATIONS.md`](https://github.com/Coder-s-OG-s/Trajectory-IR/blob/main/docs/INTEGRATIONS.md).

### When to use MCP vs CLI

| Task | Use |
|------|-----|
| Agent reads / exports / verifies inside a project | **MCP** (`trajir-mcp` + `TRAJIR_MCP_ROOT`) |
| You inspect a workdir at the terminal | **`trajirctl`** |
| CI checks a `.tir` package | **`trajirctl verify --require-signature`** |
| Browse individual nodes | **`trajirctl nodes list` / `nodes show`** (CLI only; MCP has no per node tools) |

---

## CLI command reference

```
trajirctl status       --workdir DIR --tenant ID --trajectory ID [--json]
trajirctl export       --workdir DIR --tenant ID --trajectory ID --dest PATH [--mode thin|fat] [--json]
trajirctl import       --path PATH [--src PATH] [--json]
trajirctl verify       --path PATH [--src PATH] [--require-signature] [--json]
trajirctl nodes list   --workdir DIR --tenant ID --trajectory ID [--json]
trajirctl nodes show   --workdir DIR --tenant ID --trajectory ID --id NODE_ID [--json]
```

### Flag and environment resolution

| Flag | Env fallback | Default |
|------|--------------|---------|
| `--workdir` | `TRAJIR_WORKDIR` | `.` |
| `--tenant` | `TRAJIR_TENANT` | (required) |
| `--trajectory` | `TRAJIR_TRAJECTORY` | (required) |

Flag wins over env when both are set. Package path: prefer **`--path`** (MCP name);
**`--src`** is an accepted alias on `import` and `verify`.

### Commands in detail

#### `status`

Summarize a trajectory in a workdir: paths, node count, seal count, counts by
kind.

```bash
trajirctl status --workdir ./data --tenant acme --trajectory trip-1
trajirctl status --json
```

#### `export`

Write a `.tir` package. Modes:

| Mode | Meaning |
|------|---------|
| `thin` (default) | Portable package without fatting large artifacts |
| `fat` | Include embedded artifacts as supported by the SDK |

```bash
trajirctl export --dest ./out.tir --mode thin
trajirctl export --dest ./outfat.tir --mode fat --json
```

#### `import`

Load and **report** on a `.tir` (mode, ids, node/seal counts, signed or not).
Same semantics as MCP `trajectory_import_tir`: **no NodeLog write**.

```bash
trajirctl import --path ./out.tir
trajirctl import --src ./out.tir --json    # --src alias
```

#### `verify`

Check package signature policy.

| Situation | `status` | Exit |
|-----------|----------|------|
| Unsigned, no `--require-signature` | `unsigned` | `0` |
| Signature valid | `verified` | `0` |
| Tamper / missing when required | `failed` | `1` |

```bash
trajirctl verify --path ./out.tir
trajirctl verify --path ./out.tir --require-signature   # fail closed for CI
```

#### `nodes list` / `nodes show`

Terminal native per node browsing (not exposed on MCP).

```bash
trajirctl nodes list
trajirctl nodes list --json
trajirctl nodes show --id <NODE_ID>
```

Missing `--id` prints `node not found` and exits `1`.

### Output and exit codes

- Default: human readable text on stdout
- `--json`: indented JSON with the same fields
- Errors: message on stderr

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Runtime error (including verify `failed`, nodes show not found) |
| `2` | Usage / unknown command |

---

## Development and tests

```bash
go test ./...
go vet ./...
go build -o trajirctl .
```

Tests build a fixture workdir via the public SDK (`OpenTrajectory` → `Project`
→ `SealDecision`) and exercise each command’s text and JSON paths. No
dependency on Trajectory-IR’s Python fixtures.

### Repository layout

```text
trajirctl/
  main.go                 # subcommand dispatch
  cmd/                    # status, export, import, verify, nodes
  internal/               # workdir/env resolution + text/JSON render
  integrations/mcp/       # host config templates + MCP packaging README
  scripts/                # helpers to print filled mcpServers JSON
  .github/workflows/      # CI
  LICENSE                 # Apache 2.0
```

---

## Scope and non goals

**In scope (v1):**

- Local workdir operations only
- Thin CLI over `trajir/{client,tir,workdir}`
- MCP host packaging (configs / docs), not a reimplementation of `trajir-mcp`

**Out of scope:**

- Remote / deployed backends (Postgres, Temporal, hosted API)
- A speculative `Backend` local/remote abstraction
- Changing `trajir-mcp` external behavior (lives in Trajectory-IR)
- Agent graph orchestration or replacing durable execution engines

---

## Related links

| Resource | Link |
|----------|------|
| Trajectory IR (semantics + `trajir-mcp`) | https://github.com/Coder-s-OG-s/Trajectory-IR |
| MCP integration contract | https://github.com/Coder-s-OG-s/Trajectory-IR/blob/main/docs/INTEGRATIONS.md |
| trajirctl design spec | https://github.com/Coder-s-OG-s/Trajectory-IR/blob/main/docs/superpowers/specs/2026-08-27-trajirctl-design.md |
| Go quickstart | https://github.com/Coder-s-OG-s/Trajectory-IR/blob/main/go/QUICKSTART.md |
| This repo’s MCP packaging | [`integrations/mcp/README.md`](integrations/mcp/README.md) |
| Releases | https://github.com/Coder-s-OG-s/trajirctl/releases |

---

## Contributing

1. Fork and branch from `main`
2. Keep changes focused; match existing stdlib `flag` style (no Cobra)
3. Add / update tests under `cmd/` and `internal/`
4. Run `go test ./...` and `go vet ./...` before opening a PR
5. Use conventional commits: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `ci:`

Issues and PRs welcome on
[Coder-s-OG-s/trajirctl](https://github.com/Coder-s-OG-s/trajirctl).

---

## License

Copyright 2026 Coder's OG contributors. Licensed under the
[Apache License 2.0](LICENSE).
