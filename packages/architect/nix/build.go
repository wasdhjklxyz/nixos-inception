// Package nix...(TODO)
package nix

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Build(attr string) (string, error) {
	out, err := run("nix", "build", "--print-out-paths", "--no-link", attr)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func BuildImpure(attr string, env map[string]string) error {
	cmd := exec.Command("nix", "build", "--impure", attr)

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"nix build --impure %s: %w\n%s",
			attr, err, stderr.String(),
		)
	}
	return nil
}
