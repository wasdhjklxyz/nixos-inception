// Package log...(TODO)
package log

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var useColor = term.IsTerminal(int(os.Stderr.Fd()))

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
