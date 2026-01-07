package cmd

import (
	"flag"
	"path/filepath"
	"time"
)

type flags struct {
	flake        string
	lport        int
	bootMode     string
	certDuration time.Duration
	certSkew     time.Duration
	sopsConfig   string /* FIXME: Dogshit? Use nix eval or "auto find?" */
}

func parseArgs(args []string) flags {
	f := flags{}
	fs := flag.NewFlagSet("", flag.ExitOnError)

	fs.StringVar(&f.flake, "flake", ".", "Flake configuration")

	fs.IntVar(&f.lport, "lport", -1, "Server listen port")

	fs.BoolFunc("netboot", "Use net boot", func(string) error {
		f.bootMode = "netboot"
		return nil
	})

	fs.DurationVar(
		&f.certDuration,
		"cert-duration",
		10*time.Minute,
		"Certificate validity duration",
	)

	fs.DurationVar(
		&f.certSkew,
		"cert-skew",
		5*time.Minute,
		"Certificate start time offset",
	)

	fs.Func("sops-config", "Sops configuration", func(s string) (err error) {
		f.sopsConfig, err = filepath.Abs(s)
		return
	})

	fs.Parse(args)

	return f
}
