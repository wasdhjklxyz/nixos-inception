// Package log...(TODO)
package log

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var useColor = term.IsTerminal(int(os.Stderr.Fd()))

type ProgressState struct {
	Done       int
	Running    int
	Total      int
	Bytes      int64
	TotalBytes int64
}

func Error(msg string, args ...any) {
	prefix := "error:"
	if useColor {
		prefix = "\033[1;31merror:\033[0m"
	}
	fmt.Fprintf(os.Stderr, prefix+" "+msg+"\n", args...)
}

func Warn(msg string, args ...any) {
	prefix := "warning:"
	if useColor {
		prefix = "\033[1;35mwarning:\033[0m"
	}
	fmt.Fprintf(os.Stderr, prefix+" "+msg+"\n", args...)
}

func Info(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func Highlight(msg string, args ...any) {
	if useColor {
		fmt.Fprintf(os.Stderr, "\033[1;36m"+msg+"\033[0m\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, msg+"\n", args...)
	}
}

func Progress(p ProgressState, target string) {
	mib := float64(p.Bytes) / 1024 / 1024
	totalMib := float64(p.TotalBytes) / 1024 / 1024

	const (
		reset = "\033[0m"
		green = "\033[1;92m" // completed
		blue  = "\033[1;94m" // running
		bold  = "\033[1m"    // target name
	)

	var line string
	if useColor {
		line = fmt.Sprintf(
			"\r\033[K[%s%d%s/%d copied (%s%.1f%s/%.1f MiB)] copying '%s'",
			green, p.Done, reset,
			p.Total,
			green, mib, reset,
			totalMib,
			target,
		)
	} else {
		line = fmt.Sprintf(
			"\r\033[K[%d/%d copied (%.1f/%.1f MiB)] copying '%s'",
			p.Done, p.Total, mib, totalMib, target,
		)
	}

	fmt.Fprint(os.Stderr, line)
}

func ProgressDone() {
	fmt.Fprint(os.Stderr, "\r\033[K")
}
