// Package nix...(TODO)
package nix

import (
	"strings"
)

func extractRelativePath(storePath string) string {
	const marker = "-source/"
	if idx := strings.Index(storePath, marker); idx != -1 {
		return storePath[idx+len(marker):]
	}
	return storePath
}
