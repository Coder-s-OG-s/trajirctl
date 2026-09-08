package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/tir"

	"github.com/Coder-s-OG-s/trajirctl/internal"
)

// RunImport implements `trajirctl import`: it loads and reports on a .tir
// package (no NodeLog write — same as MCP trajectory_import_tir).
// --workdir is accepted for flag-surface consistency but unused.
// --path is preferred; --src is accepted as an alias.
func RunImport(args []string, stdout io.Writer) (internal.ImportResult, error) {
	var zero internal.ImportResult
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.String("workdir", "", "unused by import; accepted for CLI consistency")
	pathFlag := fs.String("path", "", "path to the .tir package to load (preferred)")
	srcFlag := fs.String("src", "", "alias for --path")
	jsonFlag := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return zero, err
	}

	src, err := resolvePackagePath("trajirctl import", *pathFlag, *srcFlag)
	if err != nil {
		return zero, err
	}
	info, err := os.Stat(src)
	if err != nil {
		return zero, fmt.Errorf("trajirctl import: %w", err)
	}
	if !info.Mode().IsRegular() {
		return zero, fmt.Errorf("trajirctl import: %q is not a regular file", src)
	}

	pkg, err := tir.Load(src)
	if err != nil {
		return zero, fmt.Errorf("trajirctl import: %w", err)
	}

	mode, _ := pkg.Manifest["mode"].(string)
	trajectoryID, _ := pkg.Manifest["trajectory_id"].(string)
	tenantID, _ := pkg.Manifest["tenant_id"].(string)

	result := internal.ImportResult{
		Path:         src,
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
