// Package limbo...(TODO)
package limbo

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type Closure struct {
	TopLevel   string   `json:"toplevel"`
	Requisites []string `json:"requisites"`
}

func (c *Closure) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func NewClosure(topLevel string, closurePath string) (*Closure, error) {
	data, err := os.ReadFile(closurePath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	return &Closure{
		TopLevel:   topLevel,
		Requisites: lines,
	}, nil
}
