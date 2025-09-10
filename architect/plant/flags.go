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

	fs.Parse(args)

	if f.ageIdentityFile == "" {
		fs.Usage()
	}

	return f
}
