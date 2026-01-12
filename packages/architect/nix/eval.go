// Package nix...(TODO)
package nix

import (
	"encoding/json"
	"fmt"
)

func EvalJSON[T any](attr string) (T, error) {
	var result T
	out, err := run(false, "nix", "eval", "--json", attr)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return result, fmt.Errorf("parsing %s: %w", attr, err)
	}
	return result, nil
}

func EvalRaw(attr string) (string, error) {
	out, err := run(false, "nix", "eval", "--raw", attr)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func EvalApply[T any](attr string, apply string) (T, error) {
	var result T
	out, err := run(false, "nix", "eval", "--json", attr, "--apply", apply)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return result, fmt.Errorf("parsing %s: %w", attr, err)
	}
	return result, nil
}
