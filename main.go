// Command trajirctl is a human-operated CLI for local Trajectory IR
// workdirs: status, export, import, verify, and node inspection. It wraps
// the same SDK calls go/cmd/trajir-mcp exposes to agents, without that
// server's TRAJIR_MCP_ROOT confinement policy — see
// docs/superpowers/specs/2026-08-27-trajirctl-design.md §2.1 in the
// Trajectory-IR repo for why that policy doesn't apply to a terminal tool.
package main

import (
	"fmt"
	"os"

	"github.com/Coder-s-OG-s/trajirctl/cmd"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "status":
		_, err = cmd.RunStatus(os.Args[2:], os.Stdout)
	case "export":
		_, err = cmd.RunExport(os.Args[2:], os.Stdout)
	case "import":
		_, err = cmd.RunImport(os.Args[2:], os.Stdout)
	case "verify":
		_, err = cmd.RunVerify(os.Args[2:], os.Stdout)
	case "nodes":
		err = dispatchNodes(os.Args[2:])
	case "-h", "--help", "help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "trajirctl: unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func dispatchNodes(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("trajirctl nodes: expected \"list\" or \"show\"")
	}
	switch args[0] {
	case "list":
		_, err := cmd.RunNodesList(args[1:], os.Stdout)
		return err
	case "show":
		_, err := cmd.RunNodesShow(args[1:], os.Stdout)
		return err
	default:
		return fmt.Errorf("trajirctl nodes: unknown subcommand %q", args[0])
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: trajirctl <command> [flags]

commands:
  status --workdir DIR --tenant ID --trajectory ID [--json]
  export --workdir DIR --tenant ID --trajectory ID --dest PATH [--mode thin|fat] [--json]
  import --path PATH [--src PATH] [--json]
  verify --path PATH [--src PATH] [--require-signature] [--json]
  nodes list --workdir DIR --tenant ID --trajectory ID [--json]
  nodes show --workdir DIR --tenant ID --trajectory ID --id NODE_ID [--json]`)
}
