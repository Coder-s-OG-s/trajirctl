package main

import (
	"bytes"
	"crypto/ed25519"
	"path/filepath"
	"testing"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/client"
	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/tir"
)

func TestRunVerifyRequireSignatureFailsClosedAtProcessLevel(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.tir")

	// Build and export a fixture without going through run(), to keep this
	// test focused on the verify exit-code regression rather than export.
	tr, err := client.OpenTrajectory("acme", "trip-1", client.Options{
		NodesPath: filepath.Join(dir, "nodes.sqlite"),
		MemoPath:  filepath.Join(dir, "memo.sqlite"),
	})
	if err != nil {
		t.Fatalf("OpenTrajectory: %v", err)
	}
	if _, err := tr.Project(0, map[string]any{"goal": "book a flight"}); err != nil {
		t.Fatalf("Project: %v", err)
	}
	tr.Close()

	var exportOut bytes.Buffer
	if code := run([]string{
		"export", "--workdir", dir, "--tenant", "acme", "--trajectory", "trip-1", "--dest", dest,
	}, &exportOut, &exportOut); code != 0 {
		t.Fatalf("export exit code=%d, output=%s", code, exportOut.String())
	}

	// Regression test: an unsigned package under --require-signature must
	// make the *process* fail, not just print "failed" while exiting 0.
	var verifyOut, verifyErr bytes.Buffer
	code := run([]string{"verify", "--path", dest, "--require-signature"}, &verifyOut, &verifyErr)
	if code != 1 {
		t.Fatalf("exit code=%d, want 1 (stdout=%q stderr=%q)", code, verifyOut.String(), verifyErr.String())
	}

	// Signing the same package must flip the process back to success.
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if err := tir.Sign(dest, priv, tir.SignerMeta{ID: "test-signer"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}

	var signedOut bytes.Buffer
	code = run([]string{"verify", "--path", dest, "--require-signature"}, &signedOut, &signedOut)
	if code != 0 {
		t.Fatalf("exit code=%d, want 0 for a validly signed package (output=%s)", code, signedOut.String())
	}
}

func TestRunVerifyUnsignedWithoutRequireSignatureStillExitsZero(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.tir")

	tr, err := client.OpenTrajectory("acme", "trip-1", client.Options{
		NodesPath: filepath.Join(dir, "nodes.sqlite"),
		MemoPath:  filepath.Join(dir, "memo.sqlite"),
	})
	if err != nil {
		t.Fatalf("OpenTrajectory: %v", err)
	}
	if _, err := tr.Project(0, map[string]any{"goal": "book a flight"}); err != nil {
		t.Fatalf("Project: %v", err)
	}
	tr.Close()

	var out bytes.Buffer
	if code := run([]string{
		"export", "--workdir", dir, "--tenant", "acme", "--trajectory", "trip-1", "--dest", dest,
	}, &out, &out); code != 0 {
		t.Fatalf("export exit code=%d, output=%s", code, out.String())
	}

	var verifyOut bytes.Buffer
	code := run([]string{"verify", "--path", dest}, &verifyOut, &verifyOut)
	if code != 0 {
		t.Fatalf("exit code=%d, want 0 (unsigned is not a failure without --require-signature): %s", code, verifyOut.String())
	}
}

func TestRunNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("exit code=%d, want 2", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected usage on stderr")
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"bogus"}, &stdout, &stderr); code != 2 {
		t.Fatalf("exit code=%d, want 2", code)
	}
}

func TestRunHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit code=%d, want 0", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected usage on stderr")
	}
}
