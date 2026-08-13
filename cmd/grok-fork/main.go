package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/grok/fork"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func main() {
	opts := &fork.Options{
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		GrokHome: agenttty.GrokHome(),
		Env:      os.Environ(),
	}
	if err := fork.Main(os.Args[1:], opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
