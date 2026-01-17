// Package cmd...(TODO)
package cmd

import (
	"fmt"
	"os"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/dream"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/limbo"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/nix"
)

type config struct {
	addr    string
	lport   int
	netboot bool
	certDir string
}

func Run(args []string) error {
	flags := parseArgs(args)

	flake, err := nix.ResolveFlake(flags.flake)
	if err != nil {
		return fmt.Errorf("failed to resolve flake: %v", err)
	}

	cfg := mergeConfigs(flags, flake.DeployOpts)

	log.Info("generating certificates...")
	certs, err := crypto.GenerateCertificates(flags.certDuration, flags.certSkew)
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}

	log.Info("writing dreamer credentials...")
	cfg.certDir, err = dream.WriteDreamerCredentials(certs)
	if err != nil {
		return fmt.Errorf("failed to write dreamer credentials: %v", err)
	}
	defer os.RemoveAll(cfg.certDir)

	log.Info("building bootable image...")
	if err := buildDreamer(flake, cfg); err != nil {
		return fmt.Errorf("failed to build dreamer: %v", err)
	}

	log.Info("starting server...")
	if err := limbo.Descend(certs, cfg.lport, flake); err != nil {
		return fmt.Errorf("failed descent: %v", err)
	}

	return nil
}

func mergeConfigs(flags flags, deployOpts nix.DeploymentOptions) config {
	/* TODO: Remove the need for this config and netboot should be specified in
	* deployment options too for pure builds. See somewhat related issue #28 */
	cfg := config{
		addr:    deployOpts.ServerAddr,
		lport:   deployOpts.ServerPort,
		netboot: false,
	}

	if deployOpts.BootMode == "netboot" || flags.bootMode == "netboot" {
		cfg.netboot = true
	}

	return cfg
}

func buildDreamer(flake *nix.Flake, cfg config) error {
	attr := flake.ISOImage()
	if cfg.netboot {
		attr = flake.KExecTree()
	}

	env := map[string]string{"NIXOS_INCEPTION_CERT_DIR": cfg.certDir}
	if err := nix.BuildImpure(attr, env); err != nil {
		return fmt.Errorf("dreamer build failed: %v", err)
	}

	return nil
}
