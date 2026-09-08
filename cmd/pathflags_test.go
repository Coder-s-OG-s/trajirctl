package cmd

import "testing"

func TestResolvePackagePath(t *testing.T) {
	t.Parallel()

	got, err := resolvePackagePath("cmd", "/a.tir", "")
	if err != nil || got != "/a.tir" {
		t.Fatalf("path preferred: got=%q err=%v", got, err)
	}

	got, err = resolvePackagePath("cmd", "", "/b.tir")
	if err != nil || got != "/b.tir" {
		t.Fatalf("src alias: got=%q err=%v", got, err)
	}

	got, err = resolvePackagePath("cmd", "/same.tir", "/same.tir")
	if err != nil || got != "/same.tir" {
		t.Fatalf("agreeing flags: got=%q err=%v", got, err)
	}

	if _, err := resolvePackagePath("cmd", "/a.tir", "/b.tir"); err == nil {
		t.Fatal("expected disagreement error")
	}
	if _, err := resolvePackagePath("cmd", "", ""); err == nil {
		t.Fatal("expected missing path error")
	}
}
