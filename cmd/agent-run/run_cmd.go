package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agentui"
	"github.com/xhd2015/less-gen/flags"
)

const runHelp = `
Usage: agent-run run [OPTIONS] "prompt"

Options:
  --json              stream NDJSON AgentEvent lines to stdout
  --model MODEL       model name
  --session ID        session id
  --agent-runner RUNNER
  -h, --help          show help
`

func runHeadless(args []string, defaultRunner string) error {
	var jsonFlag bool
	var model string
	var sessionID string
	var agentRunner string
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--model", &model).
		String("--session", &sessionID).
		String("--agent-runner", &agentRunner).
		Help("-h,--help", runHelp).
		Parse(args)
	if err != nil {
		return err
	}
	prompt := strings.TrimSpace(strings.Join(remaining, " "))
	if prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	runner := agentRunner
	if runner == "" {
		runner = defaultRunner
	}
	if err := validateRunner(runner); err != nil {
		return err
	}
	store, err := openStore()
	if err != nil {
		return err
	}
	return agentui.Run(context.Background(), agentui.RunOptions{
		Prompt:    prompt,
		Runner:    runner,
		Model:     model,
		SessionID: sessionID,
		JSON:      jsonFlag,
		Store:     store,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	})
}