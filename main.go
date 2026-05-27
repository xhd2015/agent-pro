package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/xhd2015/agent-traces/run"
	"github.com/xhd2015/agent-traces/server"
)

//go:embed frontend/dist
var distFS embed.FS

//go:embed frontend/template.html
var templateHTML string

func main() {
	server.Init(distFS, templateHTML)

	err := run.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
