package cmd

import (
	"flag"
	"fmt"
	"io"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/client"

	"github.com/Coder-s-OG-s/trajirctl/internal"
)

func openForNodes(workDirFlag, tenantFlag, trajectoryFlag string) (tenantID, trajectoryID string, tr *client.Trajectory, err error) {
	tenantID = internal.ResolveTenant(tenantFlag)
	trajectoryID = internal.ResolveTrajectory(trajectoryFlag)
	if tenantID == "" || trajectoryID == "" {
		return "", "", nil, fmt.Errorf("--tenant and --trajectory are required")
	}
	workDir, err := internal.ResolveWorkDir(workDirFlag)
	if err != nil {
		return "", "", nil, err
	}
	nodesPath, memoPath, err := internal.SQLitePaths(workDir)
	if err != nil {
		return "", "", nil, err
	}
	tr, err = client.OpenTrajectory(tenantID, trajectoryID, client.Options{
		NodesPath: nodesPath,
		MemoPath:  memoPath,
	})
	if err != nil {
		return "", "", nil, err
	}
	return tenantID, trajectoryID, tr, nil
}

// RunNodesList implements `trajirctl nodes list`.
func RunNodesList(args []string, stdout io.Writer) (internal.NodesListResult, error) {
	var zero internal.NodesListResult
	fs := flag.NewFlagSet("nodes list", flag.ContinueOnError)
	workDirFlag := fs.String("workdir", "", "directory holding nodes.sqlite (default: TRAJIR_WORKDIR env var, else .)")
	tenantFlag := fs.String("tenant", "", "tenant id (default: TRAJIR_TENANT env var)")
	trajectoryFlag := fs.String("trajectory", "", "trajectory id (default: TRAJIR_TRAJECTORY env var)")
	jsonFlag := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return zero, err
	}

	tenantID, trajectoryID, tr, err := openForNodes(*workDirFlag, *tenantFlag, *trajectoryFlag)
	if err != nil {
		return zero, fmt.Errorf("trajirctl nodes list: %w", err)
	}
	defer tr.Close()

	rows, err := tr.Log().ListNodes(trajectoryID, tenantID)
	if err != nil {
		return zero, fmt.Errorf("trajirctl nodes list: %w", err)
	}

	summaries := make([]internal.NodeSummary, 0, len(rows))
	for _, row := range rows {
		id, _ := row["id"].(string)
		kind, _ := row["kind"].(string)
		seq, _ := row["seq"].(int)
		ts, _ := row["ts"].(float64)
		var stepN *int
		if v, ok := row["step_n"].(int); ok {
			step := v
			stepN = &step
		}
		summaries = append(summaries, internal.NodeSummary{ID: id, StepN: stepN, Seq: seq, Kind: kind, TS: ts})
	}

	result := internal.NodesListResult{TenantID: tenantID, TrajectoryID: trajectoryID, Nodes: summaries}
	if err := internal.Render(stdout, *jsonFlag, result); err != nil {
		return zero, err
	}
	return result, nil
}

// RunNodesShow implements `trajirctl nodes show`.
func RunNodesShow(args []string, stdout io.Writer) (internal.NodeShowResult, error) {
	var zero internal.NodeShowResult
	fs := flag.NewFlagSet("nodes show", flag.ContinueOnError)
	workDirFlag := fs.String("workdir", "", "directory holding nodes.sqlite (default: TRAJIR_WORKDIR env var, else .)")
	tenantFlag := fs.String("tenant", "", "tenant id (default: TRAJIR_TENANT env var)")
	trajectoryFlag := fs.String("trajectory", "", "trajectory id (default: TRAJIR_TRAJECTORY env var)")
	id := fs.String("id", "", "node id to show (required)")
	jsonFlag := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return zero, err
	}
	if *id == "" {
		return zero, fmt.Errorf("trajirctl nodes show: --id is required")
	}

	tenantID, trajectoryID, tr, err := openForNodes(*workDirFlag, *tenantFlag, *trajectoryFlag)
	if err != nil {
		return zero, fmt.Errorf("trajirctl nodes show: %w", err)
	}
	defer tr.Close()

	rows, err := tr.Log().ListNodes(trajectoryID, tenantID)
	if err != nil {
		return zero, fmt.Errorf("trajirctl nodes show: %w", err)
	}

	var result internal.NodeShowResult
	for _, row := range rows {
		if rowID, _ := row["id"].(string); rowID == *id {
			result = internal.NodeShowResult{Found: true, Node: row}
			break
		}
	}
	if err := internal.Render(stdout, *jsonFlag, result); err != nil {
		return zero, err
	}
	if !result.Found {
		return result, fmt.Errorf("trajirctl nodes show: node %q not found", *id)
	}
	return result, nil
}
