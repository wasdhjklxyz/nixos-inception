// Package nix...(TODO)
package nix

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

func (f *Flake) CheckCrossRequirements() error {
	buildSystem := getBuildSystem()
	targetSystem, err := f.getTargetSystem()
	if err != nil {
		return fmt.Errorf("failed to get flake system: %v", err)
	}

	if buildSystem == targetSystem {
		return nil
	}
	binFmtPath := filepath.Join("/proc/sys/fs/binfmt_misc", targetSystem)
	if _, err := os.Stat(binFmtPath); os.IsNotExist(err) {
		log.Warn(
			"cross-compilation from %s to %s requires binfmt emulation.\n"+
				"Add this to your NixOS config:\n"+
				"  boot.binfmt.emulatedSystems = [ \"%s\" ];",
			buildSystem, targetSystem, targetSystem,
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

func (f *Flake) getTargetSystem() (string, error) {
	out, err := EvalRaw(f.system())
	if err != nil {
		return "", err
	}
	return out, nil
}
