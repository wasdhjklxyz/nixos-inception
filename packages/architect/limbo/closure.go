// Package limbo...(TODO)
package limbo

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

type Closure struct {
	TopLevel      string   `json:"toplevel"`
	Requisites    []string `json:"requisites"`
	Disko         Disko    `json:"disko"`
	diskSelection string
}

type Disko struct {
	ScriptPath        string `json:"scriptPath"`
	PlaceholderDevice string `json:"placeholderDevice"`
	TargetDevice      string `json:"targetDevice"`
}

func (c *Closure) handler(w http.ResponseWriter, r *http.Request) {
	log.Highlight("dreamer connected from %s", r.RemoteAddr)

	var bds BlockDevices
	if err := json.NewDecoder(r.Body).Decode(&bds); err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	device, err := selectDevice(bds, c.diskSelection, c.Disko.PlaceholderDevice)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	c.Disko.TargetDevice = device

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func NewClosure(topLevel string, closurePath string, diskoScript string, diskoDevice string, diskSelection string) (*Closure, error) {
	/* TODO: Refactor. Make a "fill requisites" function instead and require
	* caller to just make their own Closure thing */
	data, err := os.ReadFile(closurePath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	return &Closure{
		TopLevel:   topLevel,
		Requisites: lines,
		Disko: Disko{
			ScriptPath:        diskoScript,
			PlaceholderDevice: diskoDevice,
		},
		diskSelection: diskSelection,
	}, nil
}
