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

	// Always install so package.json/lockfile changes are applied even when
	// node_modules already exists but is missing newer deps.
	if err := cmd.Debug().Dir("frontend-agent-run").Run("bun", "install"); err != nil {
		return err
	}

	return cmd.Debug().Dir("frontend-agent-run").Run("bun", "run", "build")
}