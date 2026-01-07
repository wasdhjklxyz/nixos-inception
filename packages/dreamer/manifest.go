package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

/* FIXME: See issue #20 this shit (struct defs) all duplicated from architect */
/* FIXME: This shit fucking sucks */

type Manifest struct {
	BlockDevices json.RawMessage `json:"blockdevices"`
	PubKey       string          `json:"pubkey"` /* NOTE: Yes ik string lolol */
}

func getBlockDevices() ([]byte, error) {
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
	return stdout.Bytes(), nil
}

func getED25519Key() (string, error) {
	str, err := os.ReadFile("/etc/ssh/ssh_host_ed25519_key.pub")
	if err != nil {
		return "", err
	}
	return string(str[12:80]), nil /* WARN: I fogot if it 80 or 81 */
}

func getManfiest() (*Manifest, error) {
	bds, err := getBlockDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get block devices: %v", err)
	}

	var bdsWrapper struct {
		BlockDevices json.RawMessage `json:"blockdevices"`
	}
	if err := json.Unmarshal(bds, &bdsWrapper); err != nil {
		return nil, err
	}

	key, err := getED25519Key()
	if err != nil {
		return nil, fmt.Errorf("failed to get pub key: %v", err)
	}
	return &Manifest{
		BlockDevices: bdsWrapper.BlockDevices,
		PubKey:       key,
	}, nil
}
