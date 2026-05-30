package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/commit_msg"
)

func main() {
	if err := commit_msg.RunGenCommitMsg(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
