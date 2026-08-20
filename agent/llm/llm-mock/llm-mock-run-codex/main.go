package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/mockpreset"
	"github.com/xhd2015/agent-pro/agent/llm/llm-mock/run"
	lessflags "github.com/xhd2015/less-flags"
)

func main() {
	opts, err := run.ParseRunFlagsFromEnv()
	if errors.Is(err, lessflags.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-codex: %v\n", err)
		os.Exit(1)
	}
	if opts.MockEventsPreset == "list" {
		mockpreset.PrintList(os.Stdout)
		return
	}
	if err := run.RunCodex(os.Args[1:], run.RunCodexOptions{
		MockEventsPreset: opts.MockEventsPreset,
		MockEventsFile:   opts.MockEventsFile,
		LogEventsPath:    opts.LogEventsPath,
		LogHTTPPath:      opts.LogHTTPPath,
		MockMCP:          opts.MockMCP,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-codex: %v\n", err)
		os.Exit(1)
	}
}
