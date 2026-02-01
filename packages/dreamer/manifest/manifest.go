// Package manifest...(TODO)
package manifest

import (
	"encoding/json"

	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/exec"
)

type Manifest struct {
	BlockDevices json.RawMessage `json:"blockdevices"`
	PubKey       string          `json:"pubkey"` /* NOTE: Age recipient */
	PrivKey      string          /* TODO: Is this serialized? */
}

func NewManifest() (*Manifest, error) {
	mf := &Manifest{}
	if err := mf.setBlockDevices(); err != nil {
		return nil, err
	}
	if err := mf.setKeys(); err != nil {
		return nil, err
	}
	return mf, nil
}

func (mf *Manifest) setBlockDevices() error {
	type LsblkOutput struct {
		BlockDevices json.RawMessage `json:"blockdevices"`
	}
	out, err := exec.RunJSON[LsblkOutput](
		"lsblk",
		"--bytes",
		"--json",
		"-o",
		"NAME,SIZE,TYPE,MODEL,PATH,RM,RO,MOUNTPOINTS",
	)
	if err != nil {
		return err
	}
	mf.BlockDevices = out.BlockDevices
	return nil
}

func (mf *Manifest) setKeys() error {
	/* FIXME: Why strings when can just use AgeKeyPair */
	kp, err := generateAgeKeyPair()
	if err != nil {
		return err
	}
	mf.PubKey = kp.recipient.String()
	mf.PrivKey = kp.identity.String()
	return nil
}
