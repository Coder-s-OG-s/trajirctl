package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/tir"

	"github.com/Coder-s-OG-s/trajirctl/internal"
)

// RunVerify implements `trajirctl verify`.
// --path is preferred; --src is accepted as an alias for MCP/CLI parity.
// Status "failed" is rendered then returned as an error so the process exits 1.
func RunVerify(args []string, stdout io.Writer) (internal.VerifyResult, error) {
	var zero internal.VerifyResult
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	pathFlag := fs.String("path", "", "path to the .tir package to verify (preferred)")
	srcFlag := fs.String("src", "", "alias for --path")
	requireSignature := fs.Bool("require-signature", false, "fail if the package is unsigned")
	jsonFlag := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return zero, err
	}
	path, err := resolvePackagePath("trajirctl verify", *pathFlag, *srcFlag)
	if err != nil {
		return zero, err
	}

	info, err := tir.Verify(path, tir.VerifyOptions{RequireSignature: *requireSignature})
	var result internal.VerifyResult
	if err != nil {
		if errors.Is(err, tir.ErrSignature) {
			result = internal.VerifyResult{
				Path:     path,
				Status:   "failed",
				Verified: false,
				Message:  err.Error(),
			}
		} else {
			return zero, fmt.Errorf("trajirctl verify: %w", err)
		}
	} else if info == nil {
		result = internal.VerifyResult{
			Path:     path,
			Status:   "unsigned",
			Signed:   false,
			Verified: false,
			Message:  "unsigned package (no SIGNATURE member); crypto verify was not run",
		}
	} else {
		scheme := ""
		if info.Document != nil {
			scheme = info.Document.Scheme
		}
		result = internal.VerifyResult{
			Path:       path,
			Status:     "verified",
			Signed:     true,
			Verified:   true,
			Scheme:     scheme,
			KeyID:      info.KeyID,
			SignerID:   info.SignerID,
			PayloadHex: fmt.Sprintf("%x", info.PayloadHash),
			Message:    "signature valid",
		}
	}

	if err := internal.Render(stdout, *jsonFlag, result); err != nil {
		return zero, err
	}
	if result.Status == "failed" {
		return result, fmt.Errorf("trajirctl verify: %s", result.Message)
	}
	return result, nil
}
