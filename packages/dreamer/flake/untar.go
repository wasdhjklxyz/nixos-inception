// Package flake...(TODO)
package flake

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const untarPath = "/tmp/flake"

func untar(tr *tar.Reader) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to advance tar archive: %v", err)
		}
		target := filepath.Join(untarPath, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("failed to make directory '%s': %v", target, err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to make directory '%s': %v", dir, err)
			}
			f, err := os.OpenFile(
				target,
				os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
				os.FileMode(hdr.Mode),
			)
			if err != nil {
				return fmt.Errorf("failed to open '%s': %v", target, err)
			}
			defer f.Close()
			if _, err := io.Copy(f, tr); err != nil {
				return fmt.Errorf("failed to copy from '%s': %v", target, err)
			}
			f.Close()
		case tar.TypeSymlink:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to make directory '%s': %v", dir, err)
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return fmt.Errorf("failed to create link '%s': %v", hdr.Linkname, err)
			}
		}
	}
	return nil
}
