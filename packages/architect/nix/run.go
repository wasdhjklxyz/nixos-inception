// Package nix...(TODO)
package nix

import (
	"bytes"
	"fmt"
	"os/exec"
)

func run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %v: %w\n%s", name, args, err, stderr.String())
	}

	return stdout.Bytes(), nil
}
