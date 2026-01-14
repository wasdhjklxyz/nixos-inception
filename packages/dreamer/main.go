package main

import (
	"archive/tar"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"filippo.io/age"
)

var (
	url    string
	client *http.Client
)

/* FIXME: Age key stuff - see #20 */

const (
	/* TODO: Strings also in lib/default.nix should use one source */
	certPath          = "/etc/nixos-inception/dreamer.crt"
	keyPath           = "/etc/nixos-inception/dreamer.key"
	caPath            = "/etc/nixos-inception/ca.crt"
	configPath        = "/etc/nixos-inception/config"
	untarPath         = "/tmp/flake"
	topLevelHeader    = "Inception-TopLevel"
	diskoScriptHeader = "Inception-DiskoScript"
)

type Disko struct {
	ScriptPath        string `json:"scriptPath"`
	PlaceholderDevice string `json:"placeholderDevice"`
	TargetDevice      string `json:"targetDevice"`
}

type Closure struct {
	TopLevel    string   `json:"toplevel"`
	Requisites  []string `json:"requisites"`
	Disko       Disko    `json:"disko"`
	SopsKeyPath string   `json:"sopskeypath"`
}

type Status struct {
	OK      bool   `json:"ok"`
	Type    string `json:"type"` /* NOTE: Yeah could be byte but whatever */
	Message string `json:"message"`
}

type AgeKeyPair struct {
	identity  *age.X25519Identity
	recipient *age.X25519Recipient
}

func generateAgeKeyPair() (*AgeKeyPair, error) {
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generated X25519 identity: %v", err)
	}
	return &AgeKeyPair{
		identity:  id,
		recipient: id.Recipient(),
	}, nil
}

func newClient() (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	caBytes, err := os.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caBytes) {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caPool,
			InsecureSkipVerify: true, /* FIXME */
			MinVersion:         tls.VersionTLS13,
		}},
		Timeout: 0, /* NOTE: No timeout TODO: Make configurable */
	}, nil
}

func buildClosure(client *http.Client, url string) (*Closure, error) {
	resp, err := client.Get(url + "/flake")
	if err != nil {
		return nil, fmt.Errorf("failed to request flake: %v", err)
	}

	topLevel := resp.Header.Get(topLevelHeader)
	if topLevel == "" {
		return nil, fmt.Errorf("missing header '%s'", topLevelHeader)
	}

	diskoScript := resp.Header.Get(diskoScriptHeader)
	if diskoScript == "" {
		return nil, fmt.Errorf("missing header '%s'", diskoScriptHeader)
	}

	if err := untarFlake(tar.NewReader(resp.Body)); err != nil {
		return nil, fmt.Errorf("failed to untar flake: %v", err)
	}

	c := &Closure{}

	c.TopLevel, err = build(untarPath + topLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to build top level: %v", err)
	}

	c.Disko.ScriptPath, err = build(untarPath + diskoScript)
	if err != nil {
		return nil, fmt.Errorf("failed to build disko script: %v", err)
	}

	if err = c.fetchClosure(client, url); err != nil {
		return nil, fmt.Errorf("failed to patch closure: %v", err)
	}

	return c, nil
}

func (c *Closure) fetchClosure(client *http.Client, url string) error {
	body, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to serialize closure: %v", err)
	}

	resp, err := client.Post(
		url+"/closure",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("failed closure request: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to deserialize closure: %v", err)
	}

	return nil
}

func getClosure(client *http.Client, url string, kp *AgeKeyPair) (*Closure, error) {
	mf, err := getManfiest(kp)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %v", err)
	}

	mfBytes, err := json.Marshal(mf)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize manifest: %v", err)
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(mfBytes))
	if err != nil {
		return nil, fmt.Errorf("failed manifest POST: %v", err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK: /* NOTE: Gotta build here */
		return buildClosure(client, url)
	case http.StatusAccepted: /* NOTE: Architect builds for us */
		var c Closure
		return &c, c.fetchClosure(client, url)
	}

	return nil, fmt.Errorf("received unexpected status: %v", resp.Status)
}

func untarFlake(tr *tar.Reader) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to advance tar archive: %v", err)
		}
		target := filepath.Join(untarPath, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("failed to make directory '%s': %v", target, err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to make directory '%s': %v", dir, err)
			}
			f, err := os.OpenFile(
				target,
				os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
				os.FileMode(hdr.Mode),
			)
			if err != nil {
				return fmt.Errorf("failed to open '%s': %v", target, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("failed to copy from '%s': %v", target, err)
			}
			f.Close()
		case tar.TypeSymlink:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("failed to make directory '%s': %v", dir, err)
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return fmt.Errorf("failed to create link '%s': %v", hdr.Linkname, err)
			}
		}
	}
	return nil
}

func build(attr string) (string, error) {
	cmd := exec.Command("nix", "build", "--print-out-paths", "--no-link", attr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
	cmd.Stderr = io.MultiWriter(&stderr, os.Stderr)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s", stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func diffClosure(requisites []string) ([]string, error) {
	ents, err := os.ReadDir("/nix/store")
	if err != nil {
		return nil, err
	}

	have := make(map[string]bool)
	for _, e := range ents {
		have[e.Name()] = true
	}

	var need []string
	for _, path := range requisites {
		name := strings.TrimPrefix(path, "/nix/store/")
		if !have[name] {
			need = append(need, name)
		}
	}
	return need, nil
}

func reportStatus(sType string, err error) {
	s := Status{OK: true, Type: sType}
	if err != nil {
		s.Message = err.Error()
	}
	sBytes, _ := json.Marshal(s)
	client.Post(url+"/status", "application/json", bytes.NewReader(sBytes))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	addr, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client, err = newClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	url = fmt.Sprintf("https://%s", strings.TrimSpace(string(addr)))

	kp, err := generateAgeKeyPair()
	reportStatus("", err)

	c, err := getClosure(client, url, kp)
	reportStatus("", err)

	need, err := diffClosure(c.Requisites)
	reportStatus("", err)

	diffReq, _ := json.Marshal(map[string][]string{"need": need})
	resp, err := client.Post(url+"/diff", "application/json", bytes.NewReader(diffReq))
	reportStatus("", err)

	var diffResp struct {
		TotalBytes int64 `json:"totalBytes"`
	}
	json.NewDecoder(resp.Body).Decode(&diffResp)
	resp.Body.Close()

	totalBytes := diffResp.TotalBytes

	for i, name := range need {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/nar/%s", url, name), nil)

		/* FIXME: Dreamer and architect to use same constants for headers */
		req.Header.Set("Inception-Total", strconv.Itoa(len(need)))
		req.Header.Set("Inception-Current", strconv.Itoa(i+1))
		req.Header.Set("Inception-Total-Bytes", strconv.FormatInt(totalBytes, 10))

		resp, err := client.Do(req)
		if err != nil {
			reportStatus("", fmt.Errorf("fetch %s: %v", name, err))
		}
		if resp.StatusCode != int(http.StatusOK) {
			reportStatus("", fmt.Errorf("fetch %s: %s", name, resp.Status))
		}

		var stderr bytes.Buffer
		cmd := exec.Command("nix-store", "--import")
		cmd.Stdin = resp.Body
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			reportStatus("", fmt.Errorf(
				"import %s: %v: %s",
				name, err, stderr.String(),
			))
		}

		resp.Body.Close()
	}

	if _, err := os.Stat(c.TopLevel); err != nil {
		reportStatus("", errors.New("top level not in store"))
	}

	_, err = client.Post(url+"/nar-done", "application/json", nil)
	reportStatus("nar", err)

	if err := os.MkdirAll("/dev/disk/by-id", 0o755); err != nil {
		reportStatus("", fmt.Errorf("failed to make disk dir(s): %v", err))
	}
	if err := os.Symlink(c.Disko.TargetDevice, c.Disko.PlaceholderDevice); err != nil {
		reportStatus("", fmt.Errorf("failed to create disk symlink: %v", err))
	}

	diskoCmd := exec.Command(c.Disko.ScriptPath)
	diskoCmd.Stdout = os.Stdout
	diskoCmd.Stderr = os.Stderr
	if err := diskoCmd.Run(); err != nil {
		reportStatus("", fmt.Errorf("disko script failed: %v", err))
	}

	installCmd := exec.Command(
		"nixos-install",
		"--no-root-passwd",
		"--system", c.TopLevel,
	)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	keyPath := filepath.Join("/mnt", c.SopsKeyPath) /* NOTE: /mnt cuz install */
	if err := os.MkdirAll(filepath.Dir(keyPath), 0o700); err != nil {
		reportStatus("", fmt.Errorf("failed to mkdir for generated key: %v", err))
	}
	if err := os.WriteFile(
		keyPath,
		[]byte(kp.identity.String()),
		0o600,
	); err != nil {
		reportStatus("", fmt.Errorf("failed to write generated key: %v", err))
	}

	reportStatus("done", installCmd.Run())
}
