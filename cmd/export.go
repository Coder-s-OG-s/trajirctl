package cmd

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/client"
	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/tir"

	"github.com/Coder-s-OG-s/trajirctl/internal"
)

// RunExport implements `trajirctl export`.
func RunExport(args []string, stdout io.Writer) (internal.ExportResult, error) {
	var zero internal.ExportResult
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	workDirFlag := fs.String("workdir", "", "directory holding nodes.sqlite (default: TRAJIR_WORKDIR env var, else .)")
	tenantFlag := fs.String("tenant", "", "tenant id (default: TRAJIR_TENANT env var)")
	trajectoryFlag := fs.String("trajectory", "", "trajectory id (default: TRAJIR_TRAJECTORY env var)")
	dest := fs.String("dest", "", "output .tir path (required)")
	modeFlag := fs.String("mode", "thin", "package mode: thin (default) or fat")
	jsonFlag := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return zero, err
	}

	tenantID := internal.ResolveTenant(*tenantFlag)
	trajectoryID := internal.ResolveTrajectory(*trajectoryFlag)
	if tenantID == "" || trajectoryID == "" {
		return zero, fmt.Errorf("trajirctl export: --tenant and --trajectory are required")
	}
	if strings.TrimSpace(*dest) == "" {
		return zero, fmt.Errorf("trajirctl export: --dest is required")
	}

	var mode tir.Mode
	switch strings.ToLower(strings.TrimSpace(*modeFlag)) {
	case "", "thin":
		mode = tir.ModeThin
	case "fat":
		mode = tir.ModeFat
	default:
		return zero, fmt.Errorf("trajirctl export: unsupported --mode %q (use thin or fat)", *modeFlag)
	}

	workDir, err := internal.ResolveWorkDir(*workDirFlag)
	if err != nil {
		return zero, fmt.Errorf("trajirctl export: %w", err)
	}
	nodesPath, memoPath, err := internal.SQLitePaths(workDir)
	if err != nil {
		return zero, fmt.Errorf("trajirctl export: %w", err)
	}

	tr, err := client.OpenTrajectory(tenantID, trajectoryID, client.Options{
		NodesPath: nodesPath,
		MemoPath:  memoPath,
	})
	if err != nil {
		return zero, fmt.Errorf("trajirctl export: %w", err)
	}
	defer tr.Close()

	path, err := tir.Export(tr.Log(), trajectoryID, *dest, tir.ExportOptions{Mode: mode})
	if err != nil {
		return zero, fmt.Errorf("trajirctl export: %w", err)
	}

	rows, err := tr.Log().ListNodes(trajectoryID, tenantID)
	if err != nil {
		return zero, fmt.Errorf("trajirctl export: %w", err)
	}

	result := internal.ExportResult{
		Path:         path,
		Mode:         string(mode),
		TrajectoryID: trajectoryID,
		TenantID:     tenantID,
		NodeCount:    len(rows),
	}
	if err := internal.Render(stdout, *jsonFlag, result); err != nil {
		return zero, err
	}
	return result, nil
}
