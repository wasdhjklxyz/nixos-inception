package cmd

import (
	"flag"
	"time"
)

type flags struct {
	flake        string
	bootMode     string
	certDuration time.Duration
	certSkew     time.Duration
}

func parseArgs(args []string) flags {
	f := flags{}
	fs := flag.NewFlagSet("", flag.ExitOnError)

	fs.StringVar(&f.flake, "flake", ".", "Flake configuration")

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

	fs.Parse(args)

	return f
}
