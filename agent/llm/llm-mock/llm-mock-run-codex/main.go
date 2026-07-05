package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/run"
)

func main() {
	if err := run.RunCodex(os.Args[1:], run.RunCodexOptions{}); err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-codex: %v\n", err)
		os.Exit(1)
	}
}