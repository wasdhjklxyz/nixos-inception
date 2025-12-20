// Package limbo...(TODO)
package limbo

import (
	"encoding/json"
	"net/http"
	"os/exec"
)

type DiffRequest struct {
	Need []string `json:"need"`
}

type DiffResponse struct {
	Paths      []string `json:"paths"`
	TotalBytes int64    `json:"totalBytes"`
}

func handleDiff(w http.ResponseWriter, r *http.Request) {
	var req DiffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	paths := make([]string, len(req.Need))
	for i, hash := range req.Need {
		paths[i] = "/nix/store/" + hash
	}

	cmd := exec.Command("nix", "path-info", "--json", "-S")
	cmd.Args = append(cmd.Args, paths...)
	out, err := cmd.Output()
	if err != nil {
		http.Error(w, "failed to get path info", http.StatusInternalServerError)
		return
	}

	var pathInfo map[string]struct {
		NarSize int64 `json:"narSize"`
	}
	json.Unmarshal(out, &pathInfo)

	var totalBytes int64
	for _, info := range pathInfo {
		totalBytes += info.NarSize
	}

	json.NewEncoder(w).Encode(DiffResponse{
		Paths:      req.Need,
		TotalBytes: totalBytes,
	})
}
