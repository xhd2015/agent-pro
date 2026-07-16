package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
		fmt.Fprintf(os.Stderr, "llm-mock-run-commandcode: %v\n", err)
		os.Exit(1)
	}
	if opts.MockEventsPreset == "list" {
		mockpreset.PrintList(os.Stdout)
		return
	}

	args := os.Args[1:]
	for _, a := range args {
		if a == "--help" || a == "-h" {
			fmt.Fprintf(os.Stderr, "llm-mock-run-commandcode: see llm-mock flags via LLM_MOCK_RUN_FLAGS=\"--help\" %s\n", strings.Join(os.Args, " "))
			break
		}
	}

	if err := run.RunCommandCode(args, run.RunCommandCodeOptions{
		MockEventsPreset: opts.MockEventsPreset,
		LogEventsPath:    opts.LogEventsPath,
		LogHTTPPath:      opts.LogHTTPPath,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "llm-mock-run-commandcode: %v\n", err)
		os.Exit(1)
	}
}
