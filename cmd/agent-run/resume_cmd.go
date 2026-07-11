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

const resumeHelp = `
Usage: agent-run resume [OPTIONS] <session-id> ["followup…"]

Resume a finished agent-run session by re-invoking the runner with the
bound provider session id (grok --resume <runner_session_id>).

Resume is a shortcut of run: same flag family, but requires the session
to be bound and the runner to have exited. If the runner is still live,
use send instead.

Arguments:
  <session-id>   agent-run session id (or runner/session_id)
  followup       follow-up prompt (required unless --open)

Options:
  --json              stream NDJSON AgentEvent lines to stdout
  --model MODEL       model name
  --keep-tty          keep TTY session alive after run completes
  --open              open keep-alive TTY and attach interactively
  --no-submit         with --open: inject prompt without trailing Enter
  --dir DIR           workspace directory (default: session workspace or cwd)
  --agent-runner RUNNER   override runner (default: from session meta)
  --agent-runner-binary SPEC
                      agent executable: bare name/path or "binary flags..."
  --agent-runner-config-home PATH
                      agent data directory (grok: GROK_HOME, codex: CODEX_HOME)
  -h, --help          show help
`

func runResume(args []string, defaultRunner string) error {
	var jsonFlag bool
	var model string
	var agentRunner string
	var agentRunnerBinary string
	var agentRunnerConfigHome string
	var keepTTY bool
	var openFlag bool
	var noSubmit bool
	var dir string
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--model", &model).
		Bool("--keep-tty", &keepTTY).
		Bool("--open", &openFlag).
		Bool("--no-submit", &noSubmit).
		String("--dir", &dir).
		String("--agent-runner", &agentRunner).
		String("--agent-runner-binary", &agentRunnerBinary).
		String("--agent-runner-config-home", &agentRunnerConfigHome).
		Help("-h,--help", resumeHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		return fmt.Errorf("resume requires <session-id>")
	}
	sessionRef := strings.TrimSpace(remaining[0])
	prompt := strings.TrimSpace(strings.Join(remaining[1:], " "))
	if prompt == "" && !openFlag {
		return fmt.Errorf("prompt is required")
	}
	if openFlag && jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if noSubmit && !openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}

	store, err := openStore()
	if err != nil {
		return err
	}
	meta, err := resolveSessionMeta(store, sessionRef)
	if err != nil {
		return err
	}

	// Gate: bound + exited.
	report := probeSessionStatus(store, meta)
	if report.Runner.Status != "bound" || strings.TrimSpace(meta.RunnerSessionID) == "" {
		return fmt.Errorf("runner session not bound (missing runner_session_id); cannot resume")
	}
	if report.Runner.Exited == nil || !*report.Runner.Exited {
		return fmt.Errorf("cannot resume: runner not exited (still active/live); use send instead of resume")
	}

	runner := strings.TrimSpace(agentRunner)
	if runner == "" {
		runner = strings.TrimSpace(meta.Runner)
	}
	if runner == "" {
		runner = defaultRunner
	}
	if err := validateRunner(runner); err != nil {
		return err
	}
	if openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}

	workspace, err := resolveRunDir(dir)
	if err != nil {
		return err
	}
	if workspace == "" {
		workspace = strings.TrimSpace(meta.Workspace)
	}
	if model == "" {
		model = strings.TrimSpace(meta.Model)
	}

	return agentui.Run(context.Background(), agentui.RunOptions{
		Prompt:                prompt,
		Runner:                runner,
		Model:                 model,
		SessionID:             meta.SessionID,
		AgentRunnerBinary:     agentRunnerBinary,
		AgentRunnerConfigHome: agentRunnerConfigHome,
		JSON:                  jsonFlag,
		Workspace:             workspace,
		KeepTerminalAlive:     keepTTY || openFlag,
		Open:                  openFlag,
		NoSubmit:              noSubmit,
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}
