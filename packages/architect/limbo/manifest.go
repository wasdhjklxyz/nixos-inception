// Package limbo...(TODO)
package limbo

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

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

	/* TODO: Make flag to choose whether to build on system or not */
	if m.flake.IsCross() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func (m *Manifest) sendFlake(w http.ResponseWriter, r *http.Request) {
	if !m.isComplete(w) {
		log.Warn("got premature flake request")
		return
	}

	w.Header().Set("Inception-TopLevel", cleanAttr(m.flake.TopLevel()))
	w.Header().Set("Inception-DiskoScript", cleanAttr(m.flake.DiskoScript()))
	w.Header().Set("Content-Type", "application/x-tar+gzip")

	gw := gzip.NewWriter(w)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	/* TODO: Warn if theres any relative paths or bring them in tar */
	if err := m.flake.Tar(tw); err != nil {
		log.Error("failed to tar flake: %v", err)
		http.Error(w, "failed to tar flake", http.StatusInternalServerError)
	}
}

func (m *Manifest) sendClosure(w http.ResponseWriter, r *http.Request) {
	if !m.isComplete(w) {
		log.Warn("got premature closure request")
		return
	}

	var c *Closure
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		log.Error("failed to deserialize closure: %v", err)
		http.Error(w, "failed to deserialize closure", http.StatusBadRequest)
		return
	}

	if c.TopLevel == "" {
		var err error
		c, err = newClosure(m.flake) /* NOTE: Builds it */
		if err != nil {
			log.Error("failed to get closure: %v", err)
			http.Error(w, "failed to get closure", http.StatusInternalServerError)
			return
		}
	}

	c.Disko.PlaceholderDevice = m.flake.DiskoDevice
	c.Disko.TargetDevice = m.targetDevice
	c.SopsKeyPath = m.flake.SopsKeyPath

	buf, err := json.Marshal(c)
	if err != nil {
		log.Error("failed to serialize closure: %v", err)
		http.Error(w, "failed to serialize closure", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func (m *Manifest) isComplete(w http.ResponseWriter) bool {
	if m.targetDevice == "" {
		http.Error(w, "incomplete manifest", http.StatusConflict)
		return false
	}
	return true
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

func cleanAttr(attr string) string {
	if idx := strings.Index(attr, "#"); idx != -1 {
		return attr[idx+1:]
	}
	return attr
}
