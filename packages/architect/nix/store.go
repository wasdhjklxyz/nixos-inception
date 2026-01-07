// Package nix...(TODO)
package nix

import (
	"strings"
)

func Requisites(paths ...string) ([]string, error) {
	args := append([]string{"-qR"}, paths...)
	out, err := run(false, "nix-store", args...)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return lines, nil
}

func extractRelativePath(storePath string) string {
	const marker = "-source/"
	if idx := strings.Index(storePath, marker); idx != -1 {
		return storePath[idx+len(marker):]
	}
	return storePath
}
