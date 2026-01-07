// Package limbo...(TODO)
package limbo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/goccy/go-yaml"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/nix"
)

type Closure struct {
	TopLevel    string   `json:"toplevel"`
	Requisites  []string `json:"requisites"`
	Disko       Disko    `json:"disko"`
	SopsKeyPath string   `json:"sopskeypath"`
	sopsConfig  string
	flake       *nix.Flake
}

type Disko struct {
	ScriptPath        string `json:"scriptPath"`
	PlaceholderDevice string `json:"placeholderDevice"`
	TargetDevice      string `json:"targetDevice"`
}

type Manifest struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
	PubKey       string        `json:"pubkey"` /* NOTE: Age recipient */
}

func (c *Closure) handler(w http.ResponseWriter, r *http.Request) {
	log.Highlight("dreamer connected from %s", r.RemoteAddr)

	var mf Manifest
	if err := json.NewDecoder(r.Body).Decode(&mf); err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	if err := updateSops(mf.PubKey, c.sopsConfig, c.flake.SopsFile); err != nil {
		http.Error(w, "sops update failed", http.StatusInternalServerError)
		return
	}

	device, err := selectDevice(
		mf.BlockDevices,
		c.flake.DeployOpts.DiskSelection,
		c.Disko.PlaceholderDevice,
	)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	c.Disko.TargetDevice = device

	log.Info("building system top level...")
	c.TopLevel, err = nix.Build(c.flake.TopLevel())
	if err != nil {
		http.Error(w, "failed top level rebuild", http.StatusInternalServerError)
		return
	}

	log.Info("building disko script...")
	c.Disko.ScriptPath, err = nix.Build(c.flake.DiskoScript())
	if err != nil {
		http.Error(w, "disko script build failed", http.StatusInternalServerError)
		return
	}

	log.Info("querying top level requisites...")
	c.Requisites, err = nix.Requisites(c.TopLevel)
	if err != nil {
		http.Error(w, "failed to query top level requisites", http.StatusInternalServerError)
		return
	}

	log.Info("querying disko script requisites...")
	reqs, err := nix.Requisites(c.Disko.ScriptPath)
	if err != nil {
		http.Error(w, "failed to get disko script requisites", http.StatusInternalServerError)
	}
	c.Requisites = append(c.Requisites, reqs...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func NewClosure(flake *nix.Flake, sopsConfig string) (*Closure, error) {
	/* FIXME: THiS IS SO FUCKING BAD */
	/* TODO: Refactor. Make a "fill requisites" function instead and require
	* caller to just make their own Closure thing */
	return &Closure{
		Disko: Disko{
			PlaceholderDevice: flake.DiskoDevice,
		},
		SopsKeyPath: flake.SopsKeyPath,
		sopsConfig:  sopsConfig,
		flake:       flake,
	}, nil
}

func updateSops(ageRecipient, sopsConfig, sopsFile string) error {
	data, err := os.ReadFile(sopsConfig)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %v", sopsConfig, err)
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse '%s': %v", sopsConfig, err)
	}

	/* NOTE: Type assertion hell but it works */
	rules := config["creation_rules"].([]any)
	rule := rules[0].(map[string]any)
	keyGroups := rule["key_groups"].([]any)
	group := keyGroups[0].(map[string]any)
	ages := group["age"].([]any)

	group["age"] = append(ages, ageRecipient)

	out, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize new sops config: %v", err)
	}

	if err := os.WriteFile(sopsConfig, out, 0o644); err != nil {
		return fmt.Errorf("failed to write new config to '%s': %v", sopsConfig, err)
	}

	cmd := exec.Command("sops", "--config", sopsConfig, "updatekeys", "-y", sopsFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update keys for '%s': %v", sopsFile, err)
	}

	return nil
}
