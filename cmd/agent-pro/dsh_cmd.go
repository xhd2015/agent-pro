package main

import (
	dshcli "github.com/xhd2015/agent-pro/agent/dsh/cli"
	"os"
)

func handleDSH(args []string) error { return dshcli.Run(args, os.Stdout, os.Stderr) }
