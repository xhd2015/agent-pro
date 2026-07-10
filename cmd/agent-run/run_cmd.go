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
  --session-id-from-prompt   generate session id from prompt slug (storage + TTY registry)
  --keep-tty          keep TTY session alive after run completes
  --agent-runner RUNNER   codex, codex-tty, grok-tty, opencode, fake-codex, ...
  --agent-runner-binary SPEC
                      agent executable: bare name/path or shell-style "binary flags..."
  --agent-runner-config-home PATH
                      agent data directory (grok: GROK_HOME, codex: CODEX_HOME);
                      default: AGENT_RUNNER_CONFIG_HOME env
  -h, --help          show help
`

func runHeadless(args []string, defaultRunner string) error {
	var jsonFlag bool
	var model string
	var sessionID string
	var sessionIDFromPrompt bool
	var agentRunner string
	var agentRunnerBinary string
	var agentRunnerConfigHome string
	var keepTTY bool
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--model", &model).
		String("--session", &sessionID).
		Bool("--session-id-from-prompt", &sessionIDFromPrompt).
		Bool("--keep-tty", &keepTTY).
		String("--agent-runner", &agentRunner).
		String("--agent-runner-binary", &agentRunnerBinary).
		String("--agent-runner-config-home", &agentRunnerConfigHome).
		Help("-h,--help", runHelp).
		Parse(args)
	if err != nil {
		return err
	}
	prompt := strings.TrimSpace(strings.Join(remaining, " "))
	if prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if sessionIDFromPrompt && strings.TrimSpace(sessionID) != "" {
		return fmt.Errorf("--session and --session-id-from-prompt are mutually exclusive; cannot use both")
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
	if sessionIDFromPrompt {
		id, genErr := generateAutoSessionID(prompt, runner, store.Home())
		if genErr != nil {
			return genErr
		}
		sessionID = id
	}
	return agentui.Run(context.Background(), agentui.RunOptions{
		Prompt:                prompt,
		Runner:                runner,
		Model:                 model,
		SessionID:             sessionID,
		AgentRunnerBinary:     agentRunnerBinary,
		AgentRunnerConfigHome: agentRunnerConfigHome,
		JSON:                  jsonFlag,
		KeepTerminalAlive:     keepTTY,
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}
