// Package flake...(TODO)
package flake

import (
	"os"
	"path/filepath"

	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/exec"
	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/manifest"
)

const copyDst = "/mnt/etc/nixos"

func (f *Flake) Install(mf *manifest.Manifest) error {
	if err := f.preInstall(mf); err != nil {
		return err
	}
	if err := f.install(); err != nil {
		return err
	}
	if err := f.postInstall(); err != nil {
		return err
	}
	return nil
}

func (f *Flake) preInstall(mf *manifest.Manifest) error {
	if err := f.Disko.RunScript(); err != nil {
		return err
	}
	if err := f.writeKey(mf); err != nil {
		return err
	}
	return nil
}

func (f *Flake) install() error {
	_, err := exec.Run(
		"nixos-install",
		"--no-root-passwd",
		"--system", f.TopLevel,
	)
	if err != nil {
		return err
	}
	return nil
}

func (f *Flake) postInstall() error {
	if err := f.copyToInstalledSystem(); err != nil {
		return err
	}
	if err := unmount(); err != nil {
		return err
	}
	return nil
}

func (f *Flake) writeKey(mf *manifest.Manifest) error {
	path := filepath.Join("/mnt", f.SopsKeyPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(mf.PrivKey), 0o600); err != nil {
		return err
	}
	return nil
}

func (f *Flake) copyToInstalledSystem() error {
	if err := os.MkdirAll(copyDst, 0o755); err != nil {
		return err
	}
	if _, err := exec.Run("cp", "-r", untarPath+"/.", copyDst); err != nil {
		return err
	}
	return nil
}

func unmount() error {
	if _, err := exec.Run("sync"); err != nil {
		return err
	}
	if _, err := exec.Run("umount", "-R", "/mnt"); err != nil {
		return err
	}
	return nil
}
