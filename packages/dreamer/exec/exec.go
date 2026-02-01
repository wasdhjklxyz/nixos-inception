// Package exec...(TODO)
package exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Result struct {
	Stdout string
	Stderr string
}

func Run(name string, args ...string) (*Result, error) {
	return run(false, name, args...)
}

func RunQuiet(name string, args ...string) (*Result, error) {
	return run(true, name, args...)
}

func RunJSON[T any](name string, args ...string) (*T, error) {
	res, err := RunQuiet(name, args...)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal([]byte(res.Stdout), &out); err != nil {
		return nil, fmt.Errorf("%s: failed to parse json: %v", name, err)
	}
	return &out, nil
}

func run(quiet bool, name string, args ...string) (*Result, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	if quiet {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	} else {
		cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
		cmd.Stderr = io.MultiWriter(&stderr, os.Stderr)
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %s", name, stderr.String())
	}
	return &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, nil
}
