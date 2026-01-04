package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	/* TODO: Strings also in lib/default.nix should use one source */
	certPath   = "/etc/nixos-inception/dreamer.crt"
	keyPath    = "/etc/nixos-inception/dreamer.key"
	caPath     = "/etc/nixos-inception/ca.crt"
	configPath = "/etc/nixos-inception/config"
)

type Disko struct {
	ScriptPath        string `json:"scriptPath"`
	PlaceholderDevice string `json:"placeholderDevice"`
	TargetDevice      string `json:"targetDevice"`
}

type Closure struct {
	TopLevel   string   `json:"toplevel"`
	Requisites []string `json:"requisites"`
	Disko      Disko    `json:"disko"`
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

func fetchClosure(client *http.Client, url string) (*Closure, error) {
	var stdout bytes.Buffer
	cmd := exec.Command(
		"lsblk",
		"--json",
		"-o",
		"NAME,SIZE,TYPE,MODEL,PATH,RM,RO,MOUNTPOINTS",
	)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("lsblk failed: %v", err)
	}
	if stdout.Len() == 0 {
		return nil, fmt.Errorf("lsblk returned empty output")
	}

	resp, err := client.Post(url, "application/json", &stdout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var c Closure
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
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

func main() {
	addr, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client, err := newClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	url := fmt.Sprintf("https://%s", strings.TrimSpace(string(addr)))

	c, err := fetchClosure(client, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	need, err := diffClosure(c.Requisites)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	diffReq, _ := json.Marshal(map[string][]string{"need": need})
	resp, err := client.Post(url+"/diff", "application/json", bytes.NewReader(diffReq))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

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
			fmt.Fprintf(os.Stderr, "fetch %s: %v\n", name, err)
			os.Exit(1)
		}
		if resp.StatusCode != int(http.StatusOK) {
			fmt.Fprintf(os.Stderr, "fetch %s: %s\n", name, resp.Status)
			os.Exit(1)
		}

		var stderr bytes.Buffer
		cmd := exec.Command("nix-store", "--import")
		cmd.Stdin = resp.Body
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "import %s: %v: %s\n", name, err, stderr.String())
			os.Exit(1)
		}

		resp.Body.Close()
	}

	if _, err := os.Stat(c.TopLevel); err != nil {
		fmt.Fprintln(os.Stderr, "top level not in store")
		os.Exit(1)
	}

	/* TODO: Should use a general status or info endpoint? */
	_, err = client.Post(url+"/nar-done", "application/json", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "POST /nar-done: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll("/dev/disk/by-id", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to make disk dir(s): %v\n", err)
		os.Exit(1)
	}
	if err := os.Symlink(c.Disko.TargetDevice, c.Disko.PlaceholderDevice); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create disk symlink: %v\n", err)
		os.Exit(1)
	}

	diskoCmd := exec.Command(c.Disko.ScriptPath)
	diskoCmd.Stdout = os.Stdout
	diskoCmd.Stderr = os.Stderr
	if err := diskoCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "disko script failed: %v\n", err)
		os.Exit(1)
	}
}
