package conn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/manifest"
)

const (
	healthEndpoint   = "/health"
	manifestEndpoint = "/"
	flakeEndpoint    = "/flake"
	closureEndpoint  = "/closure"
	statusEndpoint   = "/status"
)

const (
	/* TODO: Remove the need for these */
	flakeTopLevelHeader    = "Inception-TopLevel"
	flakeDiskoScriptHeader = "Inception-DiskoScript"
)

type FlakeMetadata struct {
	/* TODO: Remove the need for this */
	TopLevel    string
	DiskoScript string
}

func (c *Conn) GetHealth() error {
	resp, err := c.client.Get(c.url + healthEndpoint)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Conn) PostManifest(mf *manifest.Manifest) error {
	mfBytes, err := json.Marshal(mf)
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %v", err)
	}
	resp, err := c.client.Post(
		c.url+manifestEndpoint,
		"application/json",
		bytes.NewReader(mfBytes),
	)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Conn) GetFlake(w io.Writer) (fm *FlakeMetadata, err error) {
	resp, err := c.client.Get(c.url + flakeEndpoint)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	fm = &FlakeMetadata{}
	if fm.TopLevel, err = getHeader(resp, flakeTopLevelHeader); err != nil {
		return
	}
	if fm.DiskoScript, err = getHeader(resp, flakeDiskoScriptHeader); err != nil {
		return
	}
	if _, err = io.Copy(w, resp.Body); err != nil {
		return
	}
	return
}

func (c *Conn) GetClosure(v any) error {
	/* FIXME: Don't know how I feel about the any type here */
	/* FIXME: Error msgs say closure and datatype is flake */
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to serialize closure: %v", err)
	}
	resp, err := c.client.Post(
		c.url+closureEndpoint,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed closure request: %v", err)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Conn) PostStatus(statusErr error) error {
	type Status struct {
		OK      bool   `json:"ok"`
		Type    string `json:"type"` /* NOTE: Yeah could be byte but whatever */
		Message string `json:"message"`
	}
	s := Status{OK: statusErr == nil}
	if statusErr != nil {
		s.Message = statusErr.Error()
	} else {
		s.Type = "done"
	}
	sBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if _, err := c.client.Post(
		c.url+statusEndpoint,
		"application/json",
		bytes.NewReader(sBytes),
	); err != nil {
		return err
	}
	return nil
}

func getHeader(resp *http.Response, header string) (string, error) {
	val := resp.Header.Get(header)
	if val == "" {
		return "", fmt.Errorf("missing header '%s'", header)
	}
	return val, nil
}
