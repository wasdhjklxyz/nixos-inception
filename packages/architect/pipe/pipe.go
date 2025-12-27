// Package pipe...(TODO)
package pipe

import (
	"bufio"
	"fmt"
	"os"
)

type Pipe struct {
	path    string
	file    *os.File
	scanner *bufio.Scanner
}

func NewPipe(path string) (*Pipe, error) {
	f, err := os.OpenFile(path, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, fmt.Errorf("failed to open ctl pipe: %w", err)
	}
	return &Pipe{
		path:    path,
		file:    f,
		scanner: bufio.NewScanner(f),
	}, nil
}

func (p *Pipe) Close() error {
	return p.file.Close()
}

func (p *Pipe) Send(msg string) error {
	_, err := fmt.Fprintln(p.file, msg)
	return err
}

func (p *Pipe) Recv() (string, error) {
	if p.scanner.Scan() {
		return p.scanner.Text(), nil
	}
	if err := p.scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("pipe closed")
}
