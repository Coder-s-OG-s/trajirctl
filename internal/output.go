// Package internal provides output rendering and workdir/tenant/trajectory
// flag resolution shared by every trajirctl subcommand.
package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Result is implemented by every command's output struct. jsonMode picks
// between Text() and a JSON encoding of the same value — callers never
// branch on jsonMode themselves.
type Result interface {
	Text() string
}

// Render writes v to w as either its Text() form or indented JSON.
func Render(w io.Writer, jsonMode bool, v Result) error {
	if jsonMode {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	_, err := fmt.Fprintln(w, v.Text())
	return err
}

// StatusResult is the output of the status command.
type StatusResult struct {
	WorkDir      string         `json:"work_dir"`
	NodesPath    string         `json:"nodes_path"`
	TenantID     string         `json:"tenant_id"`
	TrajectoryID string         `json:"trajectory_id"`
	NodeCount    int            `json:"node_count"`
	SealCount    int            `json:"seal_count"`
	CountsByKind map[string]int `json:"counts_by_kind"`
}

func (r StatusResult) Text() string {
	return fmt.Sprintf(
		"work_dir=%s\nnodes_path=%s\ntenant_id=%s\ntrajectory_id=%s\nnode_count=%d\nseal_count=%d\ncounts_by_kind=%v",
		r.WorkDir, r.NodesPath, r.TenantID, r.TrajectoryID, r.NodeCount, r.SealCount, r.CountsByKind,
	)
}

// ExportResult is the output of the export command.
type ExportResult struct {
	Path         string `json:"path"`
	Mode         string `json:"mode"`
	TrajectoryID string `json:"trajectory_id"`
	TenantID     string `json:"tenant_id"`
	NodeCount    int    `json:"node_count"`
}

func (r ExportResult) Text() string {
	return fmt.Sprintf("exported %d node(s) to %s (mode=%s, trajectory_id=%s, tenant_id=%s)",
		r.NodeCount, r.Path, r.Mode, r.TrajectoryID, r.TenantID)
}

// ImportResult is the output of the import command.
type ImportResult struct {
	Path         string `json:"path"`
	Mode         string `json:"mode"`
	TrajectoryID string `json:"trajectory_id"`
	TenantID     string `json:"tenant_id"`
	NodeCount    int    `json:"node_count"`
	SealCount    int    `json:"seal_count"`
	Signed       bool   `json:"signed"`
}

func (r ImportResult) Text() string {
	return fmt.Sprintf("loaded %s: mode=%s trajectory_id=%s tenant_id=%s nodes=%d seals=%d signed=%v",
		r.Path, r.Mode, r.TrajectoryID, r.TenantID, r.NodeCount, r.SealCount, r.Signed)
}

// VerifyResult is the output of the verify command. Status is one of
// "unsigned", "verified", or "failed" — "unsigned" is not the same as a
// verification failure.
type VerifyResult struct {
	Path       string `json:"path"`
	Status     string `json:"status"`
	Signed     bool   `json:"signed"`
	Verified   bool   `json:"verified"`
	Scheme     string `json:"scheme,omitempty"`
	KeyID      string `json:"key_id,omitempty"`
	SignerID   string `json:"signer_id,omitempty"`
	PayloadHex string `json:"payload_hash_hex,omitempty"`
	Message    string `json:"message"`
}

func (r VerifyResult) Text() string {
	return fmt.Sprintf("%s: %s (%s)", r.Path, r.Status, r.Message)
}

// NodeSummary is one row of `nodes list` output.
type NodeSummary struct {
	ID    string  `json:"id"`
	StepN *int    `json:"step_n"`
	Seq   int     `json:"seq"`
	Kind  string  `json:"kind"`
	TS    float64 `json:"ts"`
}

// NodesListResult is the output of `nodes list`.
type NodesListResult struct {
	TenantID     string        `json:"tenant_id"`
	TrajectoryID string        `json:"trajectory_id"`
	Nodes        []NodeSummary `json:"nodes"`
}

func (r NodesListResult) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d node(s) for trajectory_id=%s tenant_id=%s\n", len(r.Nodes), r.TrajectoryID, r.TenantID)
	for _, n := range r.Nodes {
		step := "-"
		if n.StepN != nil {
			step = fmt.Sprintf("%d", *n.StepN)
		}
		fmt.Fprintf(&b, "  id=%s step=%s seq=%d kind=%s ts=%.3f\n", n.ID, step, n.Seq, n.Kind, n.TS)
	}
	return strings.TrimRight(b.String(), "\n")
}

// NodeShowResult is the output of `nodes show`.
type NodeShowResult struct {
	Found bool           `json:"found"`
	Node  map[string]any `json:"node,omitempty"`
}

func (r NodeShowResult) Text() string {
	if !r.Found {
		return "node not found"
	}
	return fmt.Sprintf("%+v", r.Node)
}
