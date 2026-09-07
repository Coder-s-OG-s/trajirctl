package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/client"
	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/tir"
)

const (
	fixtureTenant     = "acme"
	fixtureTrajectory = "trip-1"
)

// buildFixture opens a fresh trajectory in dir via the public SDK and
// appends one PROJECT_CONTEXT and one DECISION node (step 0), returning
// the two node IDs that were created (in append order) for assertions.
func buildFixture(t *testing.T, dir string) {
	t.Helper()
	tr, err := client.OpenTrajectory(fixtureTenant, fixtureTrajectory, client.Options{
		NodesPath: filepath.Join(dir, "nodes.sqlite"),
		MemoPath:  filepath.Join(dir, "memo.sqlite"),
	})
	if err != nil {
		t.Fatalf("OpenTrajectory: %v", err)
	}
	defer tr.Close()

	if _, err := tr.Project(0, map[string]any{"goal": "book a flight"}); err != nil {
		t.Fatalf("Project: %v", err)
	}
	if _, err := tr.SealDecision(0, map[string]any{"action": "search_flights"}); err != nil {
		t.Fatalf("SealDecision: %v", err)
	}
}

func TestRunStatus(t *testing.T) {
	dir := t.TempDir()
	buildFixture(t, dir)

	var buf bytes.Buffer
	result, err := RunStatus([]string{
		"--workdir", dir,
		"--tenant", fixtureTenant,
		"--trajectory", fixtureTrajectory,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if result.NodeCount != 2 {
		t.Fatalf("NodeCount=%d, want 2", result.NodeCount)
	}
	if result.SealCount != 1 {
		t.Fatalf("SealCount=%d, want 1", result.SealCount)
	}
	if result.CountsByKind["PROJECT_CONTEXT"] != 1 || result.CountsByKind["DECISION"] != 1 {
		t.Fatalf("CountsByKind=%v", result.CountsByKind)
	}
	if buf.Len() == 0 {
		t.Fatal("expected text output to be written")
	}
}

func TestRunStatusJSON(t *testing.T) {
	dir := t.TempDir()
	buildFixture(t, dir)

	var buf bytes.Buffer
	if _, err := RunStatus([]string{
		"--workdir", dir,
		"--tenant", fixtureTenant,
		"--trajectory", fixtureTrajectory,
		"--json",
	}, &buf); err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output not valid JSON: %v (%q)", err, buf.String())
	}
	if decoded["node_count"] != float64(2) {
		t.Fatalf("decoded=%v", decoded)
	}
}

func TestRunStatusRequiresTenantAndTrajectory(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	if _, err := RunStatus([]string{"--workdir", dir}, &buf); err == nil {
		t.Fatal("expected error when --tenant/--trajectory are missing")
	}
}

func TestRunExportAndImportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	buildFixture(t, dir)
	dest := filepath.Join(dir, "out.tir")

	var exportBuf bytes.Buffer
	exportResult, err := RunExport([]string{
		"--workdir", dir,
		"--tenant", fixtureTenant,
		"--trajectory", fixtureTrajectory,
		"--dest", dest,
	}, &exportBuf)
	if err != nil {
		t.Fatal(err)
	}
	if exportResult.NodeCount != 2 || exportResult.Mode != string(tir.ModeThin) {
		t.Fatalf("exportResult=%+v", exportResult)
	}

	var importBuf bytes.Buffer
	importResult, err := RunImport([]string{"--src", exportResult.Path}, &importBuf)
	if err != nil {
		t.Fatal(err)
	}
	if importResult.NodeCount != 2 {
		t.Fatalf("importResult=%+v", importResult)
	}
	if importResult.TrajectoryID != fixtureTrajectory || importResult.TenantID != fixtureTenant {
		t.Fatalf("importResult=%+v", importResult)
	}
	if importResult.Signed {
		t.Fatalf("fixture was not signed, but importResult.Signed=true")
	}
}

func TestRunVerifyUnsignedPackage(t *testing.T) {
	dir := t.TempDir()
	buildFixture(t, dir)
	dest := filepath.Join(dir, "out.tir")
	if _, err := RunExport([]string{
		"--workdir", dir,
		"--tenant", fixtureTenant,
		"--trajectory", fixtureTrajectory,
		"--dest", dest,
	}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	result, err := RunVerify([]string{"--path", dest}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "unsigned" || result.Verified {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunVerifyRequireSignatureFailsClosed(t *testing.T) {
	dir := t.TempDir()
	buildFixture(t, dir)
	dest := filepath.Join(dir, "out.tir")
	if _, err := RunExport([]string{
		"--workdir", dir,
		"--tenant", fixtureTenant,
		"--trajectory", fixtureTrajectory,
		"--dest", dest,
	}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	result, err := RunVerify([]string{"--path", dest, "--require-signature"}, &buf)
	if err != nil {
		t.Fatalf("expected a 'failed' result, not a Go error: %v", err)
	}
	if result.Status != "failed" || result.Verified {
		t.Fatalf("result=%+v", result)
	}
}

func TestRunNodesListAndShow(t *testing.T) {
	dir := t.TempDir()
	buildFixture(t, dir)

	var listBuf bytes.Buffer
	listResult, err := RunNodesList([]string{
		"--workdir", dir,
		"--tenant", fixtureTenant,
		"--trajectory", fixtureTrajectory,
	}, &listBuf)
	if err != nil {
		t.Fatal(err)
	}
	if len(listResult.Nodes) != 2 {
		t.Fatalf("Nodes=%+v", listResult.Nodes)
	}

	targetID := listResult.Nodes[0].ID
	var showBuf bytes.Buffer
	showResult, err := RunNodesShow([]string{
		"--workdir", dir,
		"--tenant", fixtureTenant,
		"--trajectory", fixtureTrajectory,
		"--id", targetID,
	}, &showBuf)
	if err != nil {
		t.Fatal(err)
	}
	if !showResult.Found {
		t.Fatalf("expected node %q to be found", targetID)
	}
	if showResult.Node["id"] != targetID {
		t.Fatalf("Node=%+v", showResult.Node)
	}
}

func TestRunNodesShowNotFound(t *testing.T) {
	dir := t.TempDir()
	buildFixture(t, dir)

	var buf bytes.Buffer
	result, err := RunNodesShow([]string{
		"--workdir", dir,
		"--tenant", fixtureTenant,
		"--trajectory", fixtureTrajectory,
		"--id", "does-not-exist",
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if result.Found {
		t.Fatalf("expected not found, got %+v", result)
	}
}
