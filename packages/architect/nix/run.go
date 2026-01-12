// Package nix...(TODO)
package nix

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func run(verbose bool, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)

	var stdout, stderr bytes.Buffer
	if verbose {
		cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
		cmd.Stderr = io.MultiWriter(&stderr, os.Stderr)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %v: %w\n%s", name, args, err, stderr.String())
	}

	return stdout.Bytes(), nil
}
