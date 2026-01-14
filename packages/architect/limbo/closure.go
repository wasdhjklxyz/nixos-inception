// Package limbo...(TODO)
package limbo

import (
	"fmt"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/nix"
)

type Closure struct {
	TopLevel    string   `json:"toplevel"`
	Requisites  []string `json:"requisites"`
	Disko       Disko    `json:"disko"`
	SopsKeyPath string   `json:"sopskeypath"`
}

type Disko struct {
	ScriptPath        string `json:"scriptPath"`
	PlaceholderDevice string `json:"placeholderDevice"`
	TargetDevice      string `json:"targetDevice"`
}

func newClosure(flake *nix.Flake, targetDevice string) (*Closure, error) {
	/* TODO: Refactor. Make a "fill requisites" function instead and require
	* caller to just make their own Closure thing */
	c := &Closure{
		Disko: Disko{
			PlaceholderDevice: flake.DiskoDevice,
			TargetDevice:      targetDevice,
		},
		SopsKeyPath: flake.SopsKeyPath,
	}

	var err error
	log.Info("building system top level...")
	c.TopLevel, err = nix.Build(flake.TopLevel())
	if err != nil {
		return nil, fmt.Errorf("failed top level build: %v", err)
	}

	log.Info("building disko script...")
	c.Disko.ScriptPath, err = nix.Build(flake.DiskoScript())
	if err != nil {
		return nil, fmt.Errorf("disko script build failed: %v", err)
	}

	log.Info("querying top level requisites...")
	c.Requisites, err = nix.Requisites(c.TopLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to query top level requisites: %v", err)
	}

	log.Info("querying disko script requisites...")
	reqs, err := nix.Requisites(c.Disko.ScriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get disko script requisites: %v", err)
	}
	c.Requisites = append(c.Requisites, reqs...)

	return c, nil
}
