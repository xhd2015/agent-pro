package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"

	// CLI still depends on agentrunapi at the binary package level (used via agentruncli).
	_ "github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func main() {
	if err := agentruncli.Handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent-run: %v\n", err)
		os.Exit(1)
	}
}
