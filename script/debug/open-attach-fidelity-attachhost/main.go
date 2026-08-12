// Command attachhost is a doctest helper: ttywatch.AttachWriter for open-attach-fidelity.
package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: attachhost <listen> <session> <mode>")
		os.Exit(2)
	}
	if _, err := ttywatch.AttachWriter(os.Args[1], os.Args[2], os.Args[3]); err != nil {
		fmt.Fprintf(os.Stderr, "attachhost: %v\n", err)
		os.Exit(1)
	}
}
