package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	if _, err := exec.LookPath("bun"); err != nil {
		return fmt.Errorf("bun is not installed, install it from https://bun.sh/docs/installation")
	}

	if _, err := os.Stat("frontend-agent-run/node_modules"); err != nil {
		if err := cmd.Debug().Dir("frontend-agent-run").Run("bun", "install"); err != nil {
			return err
		}
	}

	return cmd.Debug().Dir("frontend-agent-run").Run("bun", "run", "build")
}