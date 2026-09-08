package cmd

import (
	"fmt"
	"strings"
)

// resolvePackagePath picks --path over --src. Both empty is an error; both set
// with different values is an error. Matching MCP, --path is the preferred name.
func resolvePackagePath(cmdName, pathFlag, srcFlag string) (string, error) {
	path := strings.TrimSpace(pathFlag)
	src := strings.TrimSpace(srcFlag)
	switch {
	case path != "" && src != "" && path != src:
		return "", fmt.Errorf("%s: --path and --src disagree (%q vs %q)", cmdName, path, src)
	case path != "":
		return path, nil
	case src != "":
		return src, nil
	default:
		return "", fmt.Errorf("%s: --path is required (--src is accepted as an alias)", cmdName)
	}
}
