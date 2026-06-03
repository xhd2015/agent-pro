package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/frontend"
	"github.com/xhd2015/agent-pro/run"
	"github.com/xhd2015/agent-pro/server"
)

func main() {
	server.Init(frontend.DistFS, frontend.TemplateHTML)

	if err := run.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent-traces: %v\n", err)
		os.Exit(1)
	}
}
