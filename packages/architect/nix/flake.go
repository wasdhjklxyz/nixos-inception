// Package nix...(TODO)
package nix

import (
	"fmt"
	"strings"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

type Flake struct {
	Path        string
	Config      string
	DeployOpts  DeploymentOptions
	DiskoDevice string
	Requisites  []string
	DiskoScript string /* FIXME: Remove - only for shit limbo code */
	TopLevel    string /* FIXME: Remove - only for shit limbo code */
}

type DeploymentOptions struct {
	ServerAddr          string `json:"serverAddr"`
	ServerPort          int    `json:"serverPort"`
	AgeKeyFile          string `json:"ageKeyFile"` /* FIXME: Remove (unused) */
	BootMode            string `json:"bootMode"`   /* NOTE: "iso" | "netboot" */
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

	log.Info("querying disk info...")
	dd, err := EvalRaw(f.attr("_inception.diskoDevice"))
	if err != nil {
		return nil, fmt.Errorf("no disko device found: %v", err)
	}
	f.DiskoDevice = dd

	log.Info("querying requisites...")
	if err := f.queryRequisites(); err != nil {
		return nil, fmt.Errorf("failed to get flake requisites: %v", err)
	}

	return f, nil
}

func (f *Flake) KExecTree() string {
	return f.attr("_inception.netboot.config.system.build.kexecTree")
}

func (f *Flake) ISOImage() string {
	return f.attr("_inception.iso.config.system.build.isoImage")
}

func (f *Flake) topLevel() string {
	return f.attr("config.system.build.toplevel")
}

func (f *Flake) diskoScript() string {
	return f.attr("config.system.build.diskoScript")
}

func (f *Flake) listConfigs() ([]string, error) {
	configs, err := EvalApply[[]string](
		f.Path+"#nixosConfigurations",
		"builtins.attrNames",
	)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (f *Flake) validate() error {
	_, err := EvalApply[bool](
		f.Path+"#nixosConfigurations."+f.Config,
		"x: true",
	)
	if err != nil {
		return fmt.Errorf("configuration '%s' not found", f.Config)
	}

	_, err = EvalApply[bool](
		f.Path+"#nixosConfigurations."+f.Config+"._inception",
		"x: true",
	)
	if err != nil {
		return fmt.Errorf("configuration '%s' missing _inception module", f.Config)
	}

	return nil
}

func (f *Flake) queryRequisites() error {
	var err error
	f.TopLevel, err = Build(f.topLevel())
	if err != nil {
		return fmt.Errorf("top level build failed: %v", err)
	}

	f.DiskoScript, err = Build(f.diskoScript())
	if err != nil {
		return fmt.Errorf("disko script build failed: %v", err)
	}

	reqs, err := Requisites(f.TopLevel)
	if err != nil {
		return fmt.Errorf("failed to get top level requisites: %v", err)
	}

	diskoScriptReqs, err := Requisites(f.DiskoScript)
	if err != nil {
		return fmt.Errorf("failed to get disko script requisites: %v", err)
	}

	reqs = append(reqs, diskoScriptReqs...)
	f.Requisites = reqs
	return nil
}

func (f *Flake) attr(suffix string) string {
	return fmt.Sprintf("%s#nixosConfigurations.%s.%s", f.Path, f.Config, suffix)
}
