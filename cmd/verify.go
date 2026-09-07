package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Coder-s-OG-s/Trajectory-IR/go/trajir/tir"

	"github.com/Coder-s-OG-s/trajirctl/internal"
)

// RunVerify implements `trajirctl verify`.
func RunVerify(args []string, stdout io.Writer) (internal.VerifyResult, error) {
	var zero internal.VerifyResult
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	path := fs.String("path", "", "path to the .tir package to verify (required)")
	requireSignature := fs.Bool("require-signature", false, "fail if the package is unsigned")
	jsonFlag := fs.Bool("json", false, "emit JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return zero, err
	}
	if strings.TrimSpace(*path) == "" {
		return zero, fmt.Errorf("trajirctl verify: --path is required")
	}

	info, err := tir.Verify(*path, tir.VerifyOptions{RequireSignature: *requireSignature})
	var result internal.VerifyResult
	if err != nil {
		if errors.Is(err, tir.ErrSignature) {
			result = internal.VerifyResult{
				Path:     *path,
				Status:   "failed",
				Verified: false,
				Message:  err.Error(),
			}
		} else {
			return zero, fmt.Errorf("trajirctl verify: %w", err)
		}
	} else if info == nil {
		result = internal.VerifyResult{
			Path:     *path,
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
			Path:       *path,
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
	return result, nil
}
