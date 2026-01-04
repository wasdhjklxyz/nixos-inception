// Package limbo...(TODO)
package limbo

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
)

type DreamerStatus struct {
	OK      bool   `json:"ok"`
	Type    string `json:"type"` /* NOTE: Yeah could be byte but whatever */
	Message string `json:"message"`
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	var ds DreamerStatus
	if err := json.NewDecoder(r.Body).Decode(&ds); err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	if ds.Type == "nar" {
		log.ProgressDone()
	}

	if !ds.OK {
		done <- fmt.Errorf("%s", ds.Message)
	} else if ds.Type == "done" {
		done <- nil
	}

	w.WriteHeader(http.StatusNoContent)
}
