# trajirctl

A human-operated command-line tool for local Trajectory IR workdirs. It
wraps the same SDK calls that `go/cmd/trajir-mcp` exposes to AI agents in
the [Trajectory-IR](https://github.com/Coder-s-OG-s/Trajectory-IR) repo,
without that server's workspace-confinement policy — see that repo's
`docs/superpowers/specs/2026-08-27-trajirctl-design.md` for the design
rationale.

## Install

```
go install github.com/Coder-s-OG-s/trajirctl@latest
```

## Usage

```
trajirctl status --workdir DIR --tenant ID --trajectory ID [--json]
trajirctl export --workdir DIR --tenant ID --trajectory ID --dest PATH [--mode thin|fat] [--json]
trajirctl import --src PATH [--json]
trajirctl verify --path PATH [--require-signature] [--json]
trajirctl nodes list --workdir DIR --tenant ID --trajectory ID [--json]
trajirctl nodes show --workdir DIR --tenant ID --trajectory ID --id NODE_ID [--json]
```

`--workdir`, `--tenant`, and `--trajectory` fall back to the
`TRAJIR_WORKDIR`, `TRAJIR_TENANT`, and `TRAJIR_TRAJECTORY` environment
variables when the flag is unset; `--workdir` defaults to `.` if neither
is set.

## Scope

v1 covers local-workdir operations only — no remote/deployed backend
support. See the design spec linked above for what's deferred and why.
