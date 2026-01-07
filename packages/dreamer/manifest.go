package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

/* FIXME: See issue #20 this shit (struct defs) all duplicated from architect */
/* FIXME: This shit fucking sucks */

type Manifest struct {
	BlockDevices json.RawMessage `json:"blockdevices"`
	PubKey       string          `json:"pubkey"` /* NOTE: Age recipient */
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

func getManfiest(kp *AgeKeyPair) (*Manifest, error) {
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
	return &Manifest{
		BlockDevices: bdsWrapper.BlockDevices,
		PubKey:       kp.recipient.String(),
	}, nil
}
