package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockpreset"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/run"
)

func main() {
	args := run.PrependRunFlagsFromEnv(os.Args[1:])
	opts, remain, err := run.ParseRunFlags(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-grok: %v\n", err)
		os.Exit(1)
	}
	if opts.MockEventsPreset == "list" {
		mockpreset.PrintList(os.Stdout)
		return
	}
	if err := run.RunGrok(remain, opts); err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-grok: %v\n", err)
		os.Exit(1)
	}
}