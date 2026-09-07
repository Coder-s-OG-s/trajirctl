package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveWorkDirFlagWins(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TRAJIR_WORKDIR", filepath.Join(t.TempDir(), "unused"))
	got, err := ResolveWorkDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestResolveWorkDirFallsBackToEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TRAJIR_WORKDIR", dir)
	got, err := ResolveWorkDir("")
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestResolveWorkDirDefaultsToCwd(t *testing.T) {
	t.Setenv("TRAJIR_WORKDIR", "")
	dir := t.TempDir()
	oldwd, _ := os.Getwd()
	defer os.Chdir(oldwd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveWorkDir("")
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestResolveTenantAndTrajectory(t *testing.T) {
	t.Setenv("TRAJIR_TENANT", "env-tenant")
	t.Setenv("TRAJIR_TRAJECTORY", "env-traj")

	if got := ResolveTenant("flag-tenant"); got != "flag-tenant" {
		t.Fatalf("flag should win, got %q", got)
	}
	if got := ResolveTenant(""); got != "env-tenant" {
		t.Fatalf("should fall back to env, got %q", got)
	}
	if got := ResolveTrajectory("flag-traj"); got != "flag-traj" {
		t.Fatalf("flag should win, got %q", got)
	}
	if got := ResolveTrajectory(""); got != "env-traj" {
		t.Fatalf("should fall back to env, got %q", got)
	}
}

func TestSQLitePathsRejectsSymlinkLeaf(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	outsideDB := filepath.Join(outside, "nodes.sqlite")
	if err := os.WriteFile(outsideDB, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "nodes.sqlite")
	if err := os.Symlink(outsideDB, link); err != nil {
		t.Skipf("symlink not available: %v", err)
	}
	if _, _, err := SQLitePaths(root); err == nil {
		t.Fatal("expected symlinked nodes.sqlite to be rejected")
	}
}

func TestSQLitePathsAllowsMissingFiles(t *testing.T) {
	root := t.TempDir()
	nodesPath, memoPath, err := SQLitePaths(root)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(nodesPath) != "nodes.sqlite" || filepath.Base(memoPath) != "memo.sqlite" {
		t.Fatalf("nodesPath=%q memoPath=%q", nodesPath, memoPath)
	}
}
