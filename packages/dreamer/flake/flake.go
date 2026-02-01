// Package flake...(TODO)
package flake

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"strings"

	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/conn"
	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/exec"
)

type Flake struct {
	TopLevel    string `json:"toplevel"`
	Disko       Disko  `json:"disko"`
	SopsKeyPath string `json:"sopskeypath"`
}

func Conjure(c *conn.Conn) (f *Flake, err error) {
	md, err := initiateNetworkTransmissionRetrievalFollowedByGzipDecompressionTarballExtractionAndFilesystemPersistenceOperation(c)
	if err != nil {
		return
	}
	f = &Flake{}
	if f.TopLevel, err = build(md.TopLevel); err != nil {
		return
	}
	if f.Disko.ScriptPath, err = build(md.DiskoScript); err != nil {
		return
	}
	if err = c.GetClosure(f); err != nil { /* FIXME: Hate this function name */
		return
	}
	return
}

func initiateNetworkTransmissionRetrievalFollowedByGzipDecompressionTarballExtractionAndFilesystemPersistenceOperation(c *conn.Conn) (*conn.FlakeMetadata, error) {
	/* TODO: Remove need for conn.FlakeMetadata - maybe get this from manifest */
	var (
		md       *conn.FlakeMetadata
		fetchErr error
	)
	pr, pw := io.Pipe()
	go func() {
		md, fetchErr = c.GetFlake(pw)
		if fetchErr != nil {
			pw.CloseWithError(fetchErr)
			return
		}
		pw.Close()
	}()
	gr, err := gzip.NewReader(pr)
	if err != nil {
		pr.Close()
		return nil, err
	}
	defer gr.Close()
	if err := untar(tar.NewReader(gr)); err != nil {
		return nil, err
	}
	if fetchErr != nil {
		return nil, fetchErr
	}
	return md, nil
}

func build(attr string) (string, error) {
	r, err := exec.Run(
		"nix", "build",
		"--print-out-paths", "--no-link", "--impure",
		"--extra-experimental-features", "nix-command",
		"--extra-experimental-features", "flakes",
		untarPath+attr,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(r.Stdout), nil
}
