// Package limbo...(TODO)
package limbo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/nix"
)

type Manifest struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
	PubKey       string        `json:"pubkey"` /* NOTE: Age recipient */
	flake        *nix.Flake    `json:"-"`
	targetDevice string        `json:"-"`
}

func (m *Manifest) handler(w http.ResponseWriter, r *http.Request) {
	log.Highlight("dreamer connected from %s", r.RemoteAddr)

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Error(err.Error())
		http.Error(w, "bad request", 400)
		return
	}

	if err := updateSops(m.PubKey, m.flake.SopsFile); err != nil {
		log.Error(err.Error())
		http.Error(w, "sops update failed", http.StatusInternalServerError)
		return
	}

	device, err := selectDevice(
		m.BlockDevices,
		m.flake.DeployOpts.DiskSelection,
		m.flake.DiskoDevice,
	)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}
	m.targetDevice = device

	c, err := newClosure(m.flake, m.targetDevice)
	if err != nil {
		log.Error("failed to get closure: %v", err)
		http.Error(w, "failed to get closure", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func updateSops(ageRecipient, sopsFile string) error {
	/* TODO: Add deployment option if user is retarded for age key (master) */
	/* WARN: Overwrties this shit and DOESNT append the recipient to .sops.yaml */
	cmd := exec.Command(
		"sops", "rotate",
		"--add-age", ageRecipient,
		"--in-place", sopsFile,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"failed to update keys for '%s': %s", sopsFile, stderr.String(),
		)
	}
	return nil
}
