// Package plant
package plant

import (
	"flag"
	"path"
	"time"
)

type flags struct {
	ageIdentityFile string
	certDuration    time.Duration
	certSkew        time.Duration
	ctlPipe         string
}

func parseArgs(args []string) flags {
	f := flags{}
	fs := flag.NewFlagSet("plant", flag.ExitOnError)

	fs.Func("age-key", "Path to age identity file", func(s string) error {
		f.ageIdentityFile = path.Clean(s)
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

	fs.Func("ctl-pipe", "Path to control pipe", func(s string) error {
		f.ctlPipe = path.Clean(s)
		return nil
	})

	fs.Parse(args)

	if f.ageIdentityFile == "" || f.ctlPipe == "" {
		fs.Usage()
	}

	return f
}
