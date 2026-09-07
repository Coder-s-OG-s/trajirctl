package internal

import (
	"os"
	"path/filepath"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/workdir"
)

// ResolveWorkDir applies flag > TRAJIR_WORKDIR env var > "." (in that
// order), then canonicalizes the result to an existing directory.
func ResolveWorkDir(flagVal string) (string, error) {
	v := flagVal
	if v == "" {
		v = os.Getenv("TRAJIR_WORKDIR")
	}
	if v == "" {
		v = "."
	}
	return workdir.CanonicalizeDir(v)
}

// ResolveTenant applies flag > TRAJIR_TENANT env var.
func ResolveTenant(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	return os.Getenv("TRAJIR_TENANT")
}

// ResolveTrajectory applies flag > TRAJIR_TRAJECTORY env var.
func ResolveTrajectory(flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	return os.Getenv("TRAJIR_TRAJECTORY")
}

// SQLitePaths derives nodes.sqlite/memo.sqlite under workDir, rejecting
// either leaf if it exists as a symlink. Missing leaves are fine — the SDK
// creates them on first open.
func SQLitePaths(workDir string) (nodesPath, memoPath string, err error) {
	nodesPath = filepath.Join(workDir, "nodes.sqlite")
	memoPath = filepath.Join(workDir, "memo.sqlite")
	if err := workdir.RequireNonSymlinkLeaf(workDir, nodesPath); err != nil {
		return "", "", err
	}
	if err := workdir.RequireNonSymlinkLeaf(workDir, memoPath); err != nil {
		return "", "", err
	}
	return nodesPath, memoPath, nil
}
