package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agents/doctest/libdoc/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
