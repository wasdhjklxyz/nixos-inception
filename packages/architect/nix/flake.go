// Package nix...(TODO)
package nix

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

type Flake struct {
	Path        string
	Config      string
	DeployOpts  DeploymentOptions
	DiskoDevice string
	Requisites  []string
	SopsKeyPath string /* FIXME: I hate this */
	SopsFile    string /* FIXME: I hate this */
	System      string /* NOTE: Dreamer system architecture */
}

type DeploymentOptions struct {
	ServerAddr          string `json:"serverAddr"`
	ServerPort          int    `json:"serverPort"`
	BootMode            string `json:"bootMode"` /* NOTE: "iso" | "netboot" */
	SquashFSCompression string `json:"squashfsCompression"`
	DiskSelection       string `json:"diskSelection"` /* NOTE: "auto" | "prompt" | "specific" */
}

func ResolveFlake(attr string) (*Flake, error) {
	f := &Flake{}
	if attr == "" {
		attr = "."
	}

	if idx := strings.LastIndex(attr, "#"); idx != -1 {
		f.Path = attr[:idx]
		f.Config = attr[idx+1:]
	} else {
		f.Path = attr
	}

	if f.Config == "" {
		cfgs, err := f.listConfigs()
		if err != nil {
			return nil, fmt.Errorf("failed to get flake configurations: %v", err)
		}

		if len(cfgs) == 1 {
			f.Config = cfgs[0]
			log.Warn("using only available configuration '%s'", f.Config)
		} else if len(cfgs) == 0 {
			return nil, fmt.Errorf("no nixosConfigurations found in flake")
		} else {
			return nil, fmt.Errorf("multiple configurations found: %v", cfgs)
		}
	}

	if err := f.validate(); err != nil {
		return nil, err
	}

	do, err := EvalJSON[DeploymentOptions](f.attr("_inception.deploymentConfig"))
	if err != nil {
		log.Warn("failed to evaluate deployment config: %v", err)
	}
	f.DeployOpts = do

	skp, err := EvalRaw(f.attr("config.sops.age.keyFile"))
	if err != nil {
		log.Warn("no sops key path provided")
	}
	f.SopsKeyPath = skp

	log.Info("querying disk info...")
	dd, err := EvalRaw(f.attr("_inception.diskoDevice"))
	if err != nil {
		return nil, fmt.Errorf("no disko device found: %v", err)
	}
	f.DiskoDevice = dd

	sf, err := EvalApplyRaw(
		f.attr("config.sops.defaultSopsFile"), "builtins.toString")
	if err != nil {
		return nil, fmt.Errorf("no sops file found: %v", err)
	}
	f.SopsFile = extractRelativePath(sf)

	ds, err := EvalRaw(f.attr("config.nixpkgs.system"))
	if err != nil {
		return nil, fmt.Errorf("failed to get target system arch: %v", err)
	}
	f.System = ds

	return f, nil
}

func (f *Flake) Tar(tw *tar.Writer) error {
	return filepath.Walk(
		f.Path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			name := info.Name()
			if name == ".git" || name == "flake.lock" {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			relPath, _ := filepath.Rel(f.Path, path)
			if relPath == "." {
				return nil
			}

			if name == "flake.nix" {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				patched := patchFlakeNix(string(content))
				hdr, err := tar.FileInfoHeader(info, "")
				if err != nil {
					return err
				}
				hdr.Name = relPath
				hdr.Size = int64(len(patched))
				if err := tw.WriteHeader(hdr); err != nil {
					return err
				}
				if _, err := tw.Write([]byte(patched)); err != nil {
					return err
				}
				return nil
			}

			var link string
			if info.Mode()&os.ModeSymlink != 0 {
				link, err = os.Readlink(path)
				if err != nil {
					return err
				}
			}

			hdr, err := tar.FileInfoHeader(info, link)
			if err != nil {
				return err
			}
			hdr.Name = relPath

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}

			if !info.IsDir() {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				defer f.Close()
				io.Copy(tw, f)
			}

			return nil
		})
}

func (f *Flake) KExecTree() string {
	return f.attr("_inception.netboot.config.system.build.kexecTree")
}

func (f *Flake) ISOImage() string {
	return f.attr("_inception.iso.config.system.build.isoImage")
}

func (f *Flake) TopLevel() string {
	return f.attr("config.system.build.toplevel")
}

func (f *Flake) DiskoScript() string {
	return f.attr("config.system.build.diskoScript")
}

func (f *Flake) listConfigs() ([]string, error) {
	configs, err := EvalApplyJSON[[]string](
		f.Path+"#nixosConfigurations",
		"builtins.attrNames",
	)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (f *Flake) validate() error {
	_, err := EvalApplyJSON[bool](
		f.Path+"#nixosConfigurations."+f.Config,
		"x: true",
	)
	if err != nil {
		return fmt.Errorf("configuration '%s' not found", f.Config)
	}

	_, err = EvalApplyJSON[bool](
		f.Path+"#nixosConfigurations."+f.Config+"._inception",
		"x: true",
	)
	if err != nil {
		return fmt.Errorf("configuration '%s' missing _inception module", f.Config)
	}

	return nil
}

func (f *Flake) attr(suffix string) string {
	return fmt.Sprintf("%s#nixosConfigurations.%s.%s", f.Path, f.Config, suffix)
}

func patchFlakeNix(content string) string {
	/* NOTE: Hacky regex to remove nixos-inception specific things in the flake */
	re := regexp.MustCompile(`(?s)nixos-inception\s*=\s*\{[^}]+\};\s*`)
	content = re.ReplaceAllString(content, "")
	content = strings.ReplaceAll(
		content,
		"nixos-inception.lib.nixosSystem",
		"nixpkgs.lib.nixosSystem",
	)

	re2 := regexp.MustCompile(`,?\s*nixos-inception`)
	content = re2.ReplaceAllString(content, "")

	/* FIXME: This assumes tha tthe deployment config is in the flake.nix exactly
	* how the fucking test configuration is setup. This is bad!!!!!!! */
	re3 := regexp.MustCompile(`(?s)deployment\s*=\s*\{[^}]+\};\s*`)
	return re3.ReplaceAllString(content, "")
}
