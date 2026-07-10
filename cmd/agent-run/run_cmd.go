package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
	"github.com/xhd2015/less-gen/flags"
)

const runHelp = `
Usage: agent-run run [OPTIONS] ["prompt"]

Options:
  --json              stream NDJSON AgentEvent lines to stdout
  --model MODEL       model name
  --session ID        session id
  --session-id-from-prompt   generate session id from prompt slug (storage + TTY registry)
  --keep-tty          keep TTY session alive after run completes
  --open              open keep-alive TTY and attach interactively (silent until detach; prints session id after)
  --no-submit         with --open: inject prompt into TTY without trailing Enter (no auto-submit)
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
	var openFlag bool
	var noSubmit bool
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--model", &model).
		String("--session", &sessionID).
		Bool("--session-id-from-prompt", &sessionIDFromPrompt).
		Bool("--keep-tty", &keepTTY).
		Bool("--open", &openFlag).
		Bool("--no-submit", &noSubmit).
		String("--agent-runner", &agentRunner).
		String("--agent-runner-binary", &agentRunnerBinary).
		String("--agent-runner-config-home", &agentRunnerConfigHome).
		Help("-h,--help", runHelp).
		Parse(args)
	if err != nil {
		return err
	}
	prompt := strings.TrimSpace(strings.Join(remaining, " "))
	if prompt == "" && !openFlag {
		return fmt.Errorf("prompt is required")
	}
	if sessionIDFromPrompt && strings.TrimSpace(sessionID) != "" {
		return fmt.Errorf("--session and --session-id-from-prompt are mutually exclusive; cannot use both")
	}
	if openFlag && jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if noSubmit && !openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}
	runner := agentRunner
	if runner == "" {
		runner = defaultRunner
	}
	if err := validateRunner(runner); err != nil {
		return err
	}
	if openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
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
		KeepTerminalAlive:     keepTTY || openFlag,
		Open:                  openFlag,
		NoSubmit:              noSubmit,
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}
