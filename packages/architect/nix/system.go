// Package nix...(TODO)
package nix

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

/* NOTE: Returns true if architect running on different system than dreamer */
func (f *Flake) IsCross() bool {
	return getBuildSystem() != f.System
}

/* NOTE: Used to check requirements ONLY if cross compiling */
func (f *Flake) CheckCrossRequirements() error {
	binFmtPath := filepath.Join("/proc/sys/fs/binfmt_misc", f.System)
	if _, err := os.Stat(binFmtPath); os.IsNotExist(err) {
		log.Warn(
			"cross-compilation from %s to %s requires binfmt emulation.\n"+
				"Add this to your NixOS config:\n"+
				"  boot.binfmt.emulatedSystems = [ \"%s\" ];",
			getBuildSystem(), f.System, f.System,
		)
	}
	return nil
}

func getBuildSystem() string {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		return "x86_64-linux"
	case "arm64":
		return "aarch64-linux"
	default:
		return arch + "-linux"
	}
}
