package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/tir"

	"github.com/Coder-s-OG-s/trajirctl/internal"
)

// RunImport implements `trajirctl import`: it loads and reports on a .tir
// package. --workdir is accepted for flag-surface consistency but unused.
func RunImport(args []string, stdout io.Writer) (internal.ImportResult, error) {
	var zero internal.ImportResult
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.String("workdir", "", "unused by import; accepted for CLI consistency")
	src := fs.String("src", "", "path to the .tir package to load (required)")
	jsonFlag := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return zero, err
	}

	if strings.TrimSpace(*src) == "" {
		return zero, fmt.Errorf("trajirctl import: --src is required")
	}
	info, err := os.Stat(*src)
	if err != nil {
		return zero, fmt.Errorf("trajirctl import: %w", err)
	}
	if !info.Mode().IsRegular() {
		return zero, fmt.Errorf("trajirctl import: %q is not a regular file", *src)
	}

	pkg, err := tir.Load(*src)
	if err != nil {
		return zero, fmt.Errorf("trajirctl import: %w", err)
	}

	mode, _ := pkg.Manifest["mode"].(string)
	trajectoryID, _ := pkg.Manifest["trajectory_id"].(string)
	tenantID, _ := pkg.Manifest["tenant_id"].(string)

	result := internal.ImportResult{
		Path:         *src,
		Mode:         mode,
		TrajectoryID: trajectoryID,
		TenantID:     tenantID,
		NodeCount:    len(pkg.Nodes),
		SealCount:    len(pkg.Seals),
		Signed:       pkg.Signature != nil,
	}
	if err := internal.Render(stdout, *jsonFlag, result); err != nil {
		return zero, err
	}
	return result, nil
}
