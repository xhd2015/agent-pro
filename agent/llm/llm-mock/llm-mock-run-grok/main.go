package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/run"
)

func main() {
	if err := run.RunGrok(os.Args[1:], run.RunGrokOptions{}); err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-grok: %v\n", err)
		os.Exit(1)
	}
}