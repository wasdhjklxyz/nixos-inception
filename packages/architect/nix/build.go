// Package nix...(TODO)
package nix

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Build(attr string) (string, error) {
	out, err := run(true, "nix", "build", "--print-out-paths", "--no-link", attr)
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

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nix build --impure %s: %w", attr, err)
	}
	return nil
}
