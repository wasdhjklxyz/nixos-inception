// Package cmd...(TODO)
package cmd

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/wasdhjklxyz/nixos-inception/packages/architect/crypto"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/dream"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/limbo"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/log"
	"github.com/wasdhjklxyz/nixos-inception/packages/architect/nix"
)

type config struct {
	addr    string
	lport   int
	certDir string
}

var rootCmd = &cobra.Command{
	Use:   "architect",
	Short: "Zero-touch NixOS deployments with secrets management",
	RunE:  run,
}

var (
	flakeStr     string
	certDuration time.Duration
	certSkew     time.Duration
)

func Execute() {
	rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVar(&flakeStr, "flake", ".", "Flake configuration")
	rootCmd.Flags().DurationVar(
		&certDuration,
		"cert-duration",
		10*time.Minute,
		"Certificate validity duration",
	)
	rootCmd.Flags().DurationVar(
		&certSkew,
		"cert-skew",
		5*time.Minute,
		"Certificate start time offset",
	)
}

func run(cmd *cobra.Command, args []string) error {
	flake, err := nix.ResolveFlake(flakeStr)
	if err != nil {
		return fmt.Errorf("failed to resolve flake: %v", err)
	}

	cfg := config{
		addr:  flake.DeployOpts.ServerAddr,
		lport: flake.DeployOpts.ServerPort,
	}

	log.Info("generating certificates...")
	caState, err := crypto.InitCA(certDuration, certSkew)
	if err != nil {
		return fmt.Errorf("failed to init CA: %v", err)
	}
	serverCert, err := crypto.InitServer(
		caState,
		net.ParseIP(cfg.addr),
		certDuration,
		certSkew,
	)
	if err != nil {
		return fmt.Errorf("failed to init server: %v", err)
	}
	dreamerCert, err := crypto.IssueDreamerCert(caState, certDuration, certSkew)
	if err != nil {
		return fmt.Errorf("failed to issue dreamer cert: %v", err)
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

func buildDreamer(flake *nix.Flake, cfg config) error {
	attr := flake.Image()
	env := map[string]string{"NIXOS_INCEPTION_CERT_DIR": cfg.certDir}
	if err := nix.BuildImpure(attr, env); err != nil {
		return fmt.Errorf("dreamer build failed: %v", err)
	}
	return nil
}
