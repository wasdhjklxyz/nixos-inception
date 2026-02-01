// Package flake...(TODO)
package flake

import (
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/exec"
)

type Disko struct {
	ScriptPath        string `json:"scriptPath"`
	PlaceholderDevice string `json:"placeholderDevice"`
	TargetDevice      string `json:"targetDevice"`
}

func (d *Disko) RunScript() error {
	/* FIXME: See https://github.com/wasdhjklxyz/nixos-inception/issues/19 */
	if err := d.linkTarget(); err != nil {
		return err
	}
	if _, err := exec.Run(d.ScriptPath); err != nil {
		return err
	}
	return nil
}

func (d *Disko) linkTarget() error {
	/* FIXME: See https://github.com/wasdhjklxyz/nixos-inception/issues/19 */
	if err := os.MkdirAll("/dev/disk/by-id", 0775); err != nil {
		return err
	}
	return os.Symlink(d.TargetDevice, d.PlaceholderDevice)
}
