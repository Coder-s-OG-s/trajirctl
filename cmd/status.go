package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/client"

	"github.com/Coder-s-OG-s/trajirctl/internal"
)

// RunStatus implements `trajirctl status`. It returns the computed result
// (for tests to assert on) after already rendering it to stdout.
func RunStatus(args []string, stdout io.Writer) (internal.StatusResult, error) {
	var zero internal.StatusResult
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	workDirFlag := fs.String("workdir", "", "directory holding nodes.sqlite (default: TRAJIR_WORKDIR env var, else .)")
	tenantFlag := fs.String("tenant", "", "tenant id (default: TRAJIR_TENANT env var)")
	trajectoryFlag := fs.String("trajectory", "", "trajectory id (default: TRAJIR_TRAJECTORY env var)")
	jsonFlag := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return zero, err
	}

	tenantID := internal.ResolveTenant(*tenantFlag)
	trajectoryID := internal.ResolveTrajectory(*trajectoryFlag)
	if tenantID == "" || trajectoryID == "" {
		return zero, fmt.Errorf("trajirctl status: --tenant and --trajectory are required")
	}

	workDir, err := internal.ResolveWorkDir(*workDirFlag)
	if err != nil {
		return zero, fmt.Errorf("trajirctl status: %w", err)
	}
	nodesPath, memoPath, err := internal.SQLitePaths(workDir)
	if err != nil {
		return zero, fmt.Errorf("trajirctl status: %w", err)
	}

	tr, err := client.OpenTrajectory(tenantID, trajectoryID, client.Options{
		NodesPath: nodesPath,
		MemoPath:  memoPath,
	})
	if err != nil {
		return zero, fmt.Errorf("trajirctl status: %w", err)
	}
	defer tr.Close()

	rows, err := tr.Log().ListNodes(trajectoryID, tenantID)
	if err != nil {
		return zero, fmt.Errorf("trajirctl status: %w", err)
	}

	countsByKind := map[string]int{}
	sealCount := 0
	for _, row := range rows {
		kind, _ := row["kind"].(string)
		countsByKind[kind]++
		if kind == "DECISION" {
			sealCount++
		}
	}

	result := internal.StatusResult{
		WorkDir:      workDir,
		NodesPath:    nodesPath,
		TenantID:     tenantID,
		TrajectoryID: trajectoryID,
		NodeCount:    len(rows),
		SealCount:    sealCount,
		CountsByKind: countsByKind,
	}
	if err := internal.Render(stdout, *jsonFlag, result); err != nil {
		return zero, err
	}
	return result, nil
}
