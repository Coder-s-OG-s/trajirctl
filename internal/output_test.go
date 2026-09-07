package internal

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type fakeResult struct {
	Name string `json:"name"`
}

func (f fakeResult) Text() string {
	return "name: " + f.Name
}

func TestRenderText(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, false, fakeResult{Name: "alice"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(buf.String()); got != "name: alice" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, true, fakeResult{Name: "alice"}); err != nil {
		t.Fatal(err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output not valid JSON: %v (%q)", err, buf.String())
	}
	if decoded["name"] != "alice" {
		t.Fatalf("decoded=%v", decoded)
	}
}
