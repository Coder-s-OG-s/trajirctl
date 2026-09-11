// Command trajirctl is a human-operated CLI for local Trajectory IR
// workdirs: status, export, import, verify, and node inspection. It wraps
// the same SDK calls go/cmd/trajir-mcp exposes to agents, without that
// server's TRAJIR_MCP_ROOT confinement policy — see
// docs/superpowers/specs/2026-08-27-trajirctl-design.md §2.1 in the
// Trajectory-IR repo for why that policy doesn't apply to a terminal tool.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Coder-s-OG-s/trajirctl/cmd"
	"github.com/Coder-s-OG-s/trajirctl/internal"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches a single command and returns the process exit code. It is
// separate from main so tests can exercise exit-code behavior without
// calling os.Exit.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		usage(stderr)
		return 2
	}

	var err error
	switch args[0] {
	case "status":
		_, err = cmd.RunStatus(args[1:], stdout)
	case "export":
		_, err = cmd.RunExport(args[1:], stdout)
	case "import":
		_, err = cmd.RunImport(args[1:], stdout)
	case "verify":
		var result internal.VerifyResult
		result, err = cmd.RunVerify(args[1:], stdout)
		// RunVerify reports a failed --require-signature check via
		// result.Status, not err, so a --require-signature package that
		// is unsigned or tampered must still fail the process here.
		if err == nil && result.Status == "failed" {
			return 1
		}
	case "nodes":
		err = dispatchNodes(args[1:], stdout)
	case "-h", "--help", "help":
		usage(stderr)
		return 0
	default:
		fmt.Fprintf(stderr, "trajirctl: unknown command %q\n", args[0])
		usage(stderr)
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func dispatchNodes(args []string, stdout io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("trajirctl nodes: expected \"list\" or \"show\"")
	}
	switch args[0] {
	case "list":
		_, err := cmd.RunNodesList(args[1:], stdout)
		return err
	case "show":
		_, err := cmd.RunNodesShow(args[1:], stdout)
		return err
	default:
		return fmt.Errorf("trajirctl nodes: unknown subcommand %q", args[0])
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, `usage: trajirctl <command> [flags]

commands:
  status --workdir DIR --tenant ID --trajectory ID [--json]
  export --workdir DIR --tenant ID --trajectory ID --dest PATH [--mode thin|fat] [--json]
  import --path PATH [--src PATH] [--json]
  verify --path PATH [--src PATH] [--require-signature] [--json]
  nodes list --workdir DIR --tenant ID --trajectory ID [--json]
  nodes show --workdir DIR --tenant ID --trajectory ID --id NODE_ID [--json]`)
}
