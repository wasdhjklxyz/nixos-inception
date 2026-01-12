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
