package main

import (
	"errors"
	"fmt"
	"os"

	lessflags "github.com/xhd2015/less-flags"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockpreset"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/run"
)

func main() {
	opts, err := run.ParseRunFlagsFromEnv()
	if errors.Is(err, lessflags.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-grok: %v\n", err)
		os.Exit(1)
	}
	if opts.MockEventsPreset == "list" {
		mockpreset.PrintList(os.Stdout)
		return
	}
	if err := run.RunGrok(os.Args[1:], opts); err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-grok: %v\n", err)
		os.Exit(1)
	}
}