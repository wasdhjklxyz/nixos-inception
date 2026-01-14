// Package nix...(TODO)
package nix

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

func GetBuildSystem() string {
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

func (f *Flake) CheckCrossRequirements(buildSystem string) error {
	targetSystem, err := f.getTargetSystem()
	if err != nil {
		return fmt.Errorf("failed to get flake system: %v", err)
	}

	if buildSystem == targetSystem {
		return nil
	}
	binFmtPath := filepath.Join(
		"/proc/sys/fs/binfmt_misc",
		archToBinfmt(targetSystem),
	)
	if _, err := os.Stat(binFmtPath); os.IsNotExist(err) {
		log.Warn(
			"cross-compilation from %s to %s requires binfmt emulation.\n"+
				"Add this to your NixOS config:\n\n"+
				"  boot.binfmt.emulatedSystems = [ \"%s\" ];",
			buildSystem, targetSystem, targetSystem,
		)
	}
	return nil
}

func (f *Flake) getTargetSystem() (string, error) {
	out, err := EvalRaw(f.system())
	if err != nil {
		return "", err
	}
	return out, nil
}

func archToBinfmt(system string) string {
	switch system {
	case "aarch64-linux":
		return "qemu-aarch64"
	case "armv7l-linux":
		return "qemu-arm"
	case "riscv64-linux":
		return "qemu-riscv64"
	case "x86_64-linux":
		return "qemu-x86_64"
	case "i686-linux":
		return "qemu-i386"
	default:
		return ""
	}
}
