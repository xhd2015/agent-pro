package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/llm/mock-mcp/run"
)

func main() {
	if err := run.Main(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
