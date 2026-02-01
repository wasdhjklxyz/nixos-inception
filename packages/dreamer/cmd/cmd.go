// Package cmd...(TODO)
package cmd

import (
	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/conn"
	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/flake"
	"github.com/wasdhjklxyz/nixos-inception/packages/dreamer/manifest"
)

func Run() (err error) {
	conn, err := conn.NewConn()
	if err != nil {
		return err
	}
	defer func() {
		conn.PostStatus(err)
	}()
	mf, err := manifest.NewManifest()
	if err != nil {
		return err
	}
	if err = conn.PostManifest(mf); err != nil {
		return err
	}
	flake, err := flake.Conjure(conn)
	if err != nil {
		return err
	}
	if err = flake.Install(mf); err != nil {
		return err
	}
	return nil
}
