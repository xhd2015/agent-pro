package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/explain"
)

func main() {
	if err := explain.RunExplain(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
