package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/run"
)

func main() {
	if err := run.RunOpencode(os.Args[1:], run.RunOpencodeOptions{}); err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-opencode: %v\n", err)
		os.Exit(1)
	}
}