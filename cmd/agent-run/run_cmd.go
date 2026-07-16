package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
	"github.com/xhd2015/agent-pro/pkgs/shell"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	flags "github.com/xhd2015/less-flags"
)

const runHelp = `
Usage: agent-run run [OPTIONS] ["prompt"]

Options:
  --json              stream NDJSON AgentEvent lines to stdout
  --model MODEL       model name
  --session ID        session id
  --session-id ID     alias for --session
  --session-id-from-prompt   generate session id from prompt slug (storage + TTY registry)
  --auto-send-or-resume      with --session-id: live→send, exited+bound→resume, else→run
  --new-terminal      with --auto-send-or-resume: open a new iTerm2 window for run/resume
                      (ignored when MODE=send / live)
  --keep-tty          keep TTY session alive after run completes
  --open              open keep-alive TTY and attach interactively (silent until detach; prints session id after);
                      with --auto-send-or-resume: required shape for create/resume; ignored when live (send only)
  --detach            start keep-alive TTY daemon and exit after registry (+ soft grok bind);
                      prints session-id and terminal-id on stdout; no attach / no stream;
                      exclusive with --open and --json; TTY only; empty prompt OK
  --no-submit         with --open: inject prompt into TTY without trailing Enter (no auto-submit);
                      with --auto-send-or-resume send path: inject without Enter
  --dir DIR           workspace directory (default: process cwd; resume uses session workspace)
  --allow-relocate-resume-session-dir
                      when resume --dir differs from grok session cwd, relocate the
                      grok session and continue (grok-tty only; with --auto-send-or-resume)
  --agent-runner RUNNER   codex, codex-tty, grok-tty, opencode, fake-codex, ...
  --agent-runner-binary SPEC
                      agent executable: bare name/path or shell-style "binary flags..."
  --agent-runner-config-home PATH
                      agent data directory (grok: GROK_HOME, codex: CODEX_HOME);
                      default: AGENT_RUNNER_CONFIG_HOME env
  --prepend-path DIR  prepend DIR to the TTY agent runner child PATH (repeatable; TTY only)
  -e, --env KEY=VALUE set env var on the TTY agent runner child process (repeatable; TTY only)
  -h, --help          show help
`

func runHeadless(args []string, defaultRunner string) error {
	var jsonFlag bool
	var model string
	var sessionID string
	var sessionIDFromPrompt bool
	var autoSendOrResume bool
	var newTerminal bool
	var agentRunner string
	var agentRunnerBinary string
	var agentRunnerConfigHome string
	var prependPaths []string
	var envEntries []string
	var keepTTY bool
	var openFlag bool
	var detachFlag bool
	var noSubmit bool
	var dir string
	var allowRelocateResumeSessionDir bool
	var recorded flags.Flags
	remaining, err := flags.Bool("--json", &jsonFlag).
		String("--model", &model).
		String("--session,--session-id", &sessionID).
		Bool("--session-id-from-prompt", &sessionIDFromPrompt).
		Bool("--auto-send-or-resume", &autoSendOrResume).
		Bool("--new-terminal", &newTerminal).
		Bool("--keep-tty", &keepTTY).
		Bool("--open", &openFlag).
		Bool("--detach", &detachFlag).
		Bool("--no-submit", &noSubmit).
		String("--dir", &dir).
		Bool("--allow-relocate-resume-session-dir", &allowRelocateResumeSessionDir).
		String("--agent-runner", &agentRunner).
		String("--agent-runner-binary", &agentRunnerBinary).
		String("--agent-runner-config-home", &agentRunnerConfigHome).
		StringSlice("--prepend-path", &prependPaths).
		StringSlice("-e,--env", &envEntries).
		Help("-h,--help", runHelp).
		CollectParsedFlags(&recorded).
		Parse(args)
	if err != nil {
		return err
	}
	prompt := strings.TrimSpace(strings.Join(remaining, " "))

	if err := validateEnvFlags(envEntries); err != nil {
		return err
	}
	absPrepend, err := resolvePrependPaths(prependPaths)
	if err != nil {
		return err
	}
	absConfigHome, err := resolveAgentRunnerConfigHomeAbs(agentRunnerConfigHome)
	if err != nil {
		return err
	}
	envEntries = normalizeEnvEntries(envEntries)

	if newTerminal && !autoSendOrResume {
		return fmt.Errorf("--new-terminal requires --auto-send-or-resume")
	}

	if autoSendOrResume {
		return runAutoSendOrResume(autoSendOrResumeOpts{
			jsonFlag:                      jsonFlag,
			model:                         model,
			sessionID:                     sessionID,
			sessionIDFromPrompt:           sessionIDFromPrompt,
			agentRunner:                   agentRunner,
			agentRunnerBinary:             agentRunnerBinary,
			agentRunnerConfigHome:         absConfigHome,
			prependPaths:                  absPrepend,
			envEntries:                    envEntries,
			keepTTY:                       keepTTY,
			openFlag:                      openFlag,
			detachFlag:                    detachFlag,
			noSubmit:                      noSubmit,
			dir:                           dir,
			allowRelocateResumeSessionDir: allowRelocateResumeSessionDir,
			prompt:                        prompt,
			defaultRunner:                 defaultRunner,
			newTerminal:                   newTerminal,
			recorded:                      recorded,
		})
	}

	if detachFlag && openFlag {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if detachFlag && jsonFlag {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if openFlag && jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if noSubmit && !openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}
	if prompt == "" && !openFlag && !detachFlag {
		return fmt.Errorf("prompt is required")
	}
	if sessionIDFromPrompt && strings.TrimSpace(sessionID) != "" {
		return fmt.Errorf("--session/--session-id and --session-id-from-prompt are mutually exclusive; cannot use both")
	}
	workspace, err := resolveRunDir(dir)
	if err != nil {
		return err
	}
	runner := resolveCLIRunner(agentRunner, defaultRunner)
	if err := validateRunner(runner); err != nil {
		return err
	}
	if openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if detachFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--detach requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if err := requireTTYForSessionEnv(runner, absPrepend, envEntries); err != nil {
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
		AgentRunnerConfigHome: absConfigHome,
		PrependPaths:          absPrepend,
		Env:                   envEntries,
		JSON:                  jsonFlag,
		Workspace:             workspace,
		KeepTerminalAlive:     keepTTY || openFlag || detachFlag,
		Open:                  openFlag,
		Detach:                detachFlag,
		NoSubmit:              noSubmit,
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}

// resolveCLIRunner picks the agent runner for run/auto paths.
// Explicit --agent-runner wins, then global default, then grok-tty (agent-run's
// primary TTY runner). agentui still falls back to store config / opencode when
// Runner is left empty by other call sites.
func resolveCLIRunner(flagRunner, defaultRunner string) string {
	if r := strings.TrimSpace(flagRunner); r != "" {
		return r
	}
	if r := strings.TrimSpace(defaultRunner); r != "" {
		return r
	}
	return "grok-tty"
}

type autoSendOrResumeOpts struct {
	jsonFlag                      bool
	model                         string
	sessionID                     string
	sessionIDFromPrompt           bool
	agentRunner                   string
	agentRunnerBinary             string
	agentRunnerConfigHome         string
	prependPaths                  []string
	envEntries                    []string
	keepTTY                       bool
	openFlag                      bool
	detachFlag                    bool
	noSubmit                      bool
	dir                           string
	allowRelocateResumeSessionDir bool
	prompt                        string
	defaultRunner                 string
	newTerminal                   bool
	recorded                      flags.Flags
}

// runAutoSendOrResume classifies a stable --session-id into MODE=run|send|resume
// and dispatches existing run/send/resume semantics.
func runAutoSendOrResume(opts autoSendOrResumeOpts) error {
	if opts.sessionIDFromPrompt {
		return fmt.Errorf("--auto-send-or-resume and --session-id-from-prompt are mutually exclusive; cannot use both")
	}
	sessionID := strings.TrimSpace(opts.sessionID)
	if sessionID == "" {
		return fmt.Errorf("--auto-send-or-resume requires --session-id")
	}
	if opts.detachFlag && opts.openFlag {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if opts.detachFlag && opts.jsonFlag {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if opts.openFlag && opts.jsonFlag {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}

	store, err := openStore()
	if err != nil {
		return err
	}

	meta, found, err := tryResolveSessionMeta(store, sessionID)
	if err != nil {
		return err
	}

	mode := "run"
	if found {
		report := probeSessionStatus(store, meta)
		if report.Resume.Ready {
			mode = "resume"
		} else if report.Runner.Exited != nil && !*report.Runner.Exited {
			mode = "send"
		}
	}

	switch mode {
	case "send":
		return autoSendLive(store, meta, opts)
	case "resume":
		if opts.newTerminal {
			return openAutoInNewTerminal(opts, meta, found)
		}
		return resumeExistingSession(store, meta, resumeRunConfig{
			jsonFlag:                      opts.jsonFlag,
			model:                         opts.model,
			agentRunner:                   opts.agentRunner,
			agentRunnerBinary:             opts.agentRunnerBinary,
			agentRunnerConfigHome:         opts.agentRunnerConfigHome,
			prependPaths:                  opts.prependPaths,
			envEntries:                    opts.envEntries,
			keepTTY:                       opts.keepTTY,
			openFlag:                      opts.openFlag,
			detachFlag:                    opts.detachFlag,
			noSubmit:                      opts.noSubmit,
			dir:                           opts.dir,
			allowRelocateResumeSessionDir: opts.allowRelocateResumeSessionDir,
			prompt:                        opts.prompt,
			defaultRunner:                 opts.defaultRunner,
		})
	default:
		if opts.newTerminal {
			return openAutoInNewTerminal(opts, meta, found)
		}
		return autoRunCreate(store, sessionID, opts)
	}
}

// openAutoInNewTerminal re-launches the auto-send-or-resume command in a new
// iTerm2 window (ModeForceNew), stripping --new-terminal so the child runs in-process.
// The launcher does not spawn the provider.
func openAutoInNewTerminal(opts autoSendOrResumeOpts, meta agentstorage.SessionMeta, found bool) error {
	childFlags := opts.recorded.Remove("--new-terminal")
	tokens := childFlags.Reconstruct()
	argv := make([]string, 0, 1+len(tokens)+2)
	argv = append(argv, "run")
	argv = append(argv, tokens...)
	if opts.prompt != "" {
		argv = append(argv, "--", opts.prompt)
	}

	exe, err := agentRunExecutable()
	if err != nil {
		return err
	}
	quoted := make([]string, 0, 1+len(argv))
	quoted = append(quoted, shell.ShellQuote(exe))
	for _, tok := range argv {
		quoted = append(quoted, shell.ShellQuote(tok))
	}
	followUp := strings.Join(quoted, " ")

	dir, err := resolveNewTerminalDir(opts.dir, meta, found)
	if err != nil {
		return err
	}
	return iterm2.OpenConfig(dir, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{followUp},
	})
}

// agentRunExecutable returns an absolute path to this process binary for re-exec.
func agentRunExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		if len(os.Args) > 0 && os.Args[0] != "" {
			exe = os.Args[0]
		} else {
			return "", fmt.Errorf("resolve agent-run executable: %w", err)
		}
	}
	if abs, absErr := filepath.Abs(exe); absErr == nil {
		exe = abs
	}
	if resolved, evalErr := filepath.EvalSymlinks(exe); evalErr == nil {
		exe = resolved
	}
	return exe, nil
}

// resolveNewTerminalDir picks the workspace for iterm2.OpenConfig:
// --dir if set, else session meta workspace (when found), else process cwd.
func resolveNewTerminalDir(dir string, meta agentstorage.SessionMeta, found bool) (string, error) {
	if resolved, err := resolveRunDir(dir); err != nil {
		return "", err
	} else if resolved != "" {
		return resolved, nil
	}
	if found {
		if w := strings.TrimSpace(meta.Workspace); w != "" {
			abs, err := filepath.Abs(w)
			if err != nil {
				return "", fmt.Errorf("workspace: %w", err)
			}
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				if real, err := filepath.EvalSymlinks(abs); err == nil {
					return real, nil
				}
				return abs, nil
			}
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cwd: %w", err)
	}
	return cwd, nil
}

func autoSendLive(store agentstorage.Store, meta agentstorage.SessionMeta, opts autoSendOrResumeOpts) error {
	// --open/--detach are accepted but ignored while live: callers (e.g. local-bot)
	// always pass the same CLI for run/send/resume. Live work is enqueue + wait delivery only.
	if opts.openFlag {
		fmt.Fprintln(os.Stderr, "note: --open ignored while session is live; sending follow-up")
	}
	if opts.detachFlag {
		fmt.Fprintln(os.Stderr, "note: --detach ignored while session is live; sending follow-up")
	}
	if opts.newTerminal {
		fmt.Fprintln(os.Stderr, "note: --new-terminal ignored while session is live; sending follow-up")
	}
	if opts.prompt == "" {
		fmt.Fprintln(os.Stderr, "warning: session is live; no message to send")
		return nil
	}

	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = strings.TrimSpace(meta.SessionID)
	}
	if termID == "" {
		return fmt.Errorf("cannot send: missing terminal_session_id and session_id")
	}

	home := store.Home()
	ttySess, err := agenttty.ResolveByTerminalID(home, termID)
	if err != nil {
		return err
	}
	if !ttySess.TCPReachable {
		return fmt.Errorf("terminal unreachable at %s", ttySess.Registry.ListenAddr)
	}
	provider, ok := agenttty.Get(ttySess.RunnerID)
	if !ok {
		return fmt.Errorf("unknown tty runner: %s", ttySess.RunnerID)
	}

	sess := agentsend.Session{
		Home:              home,
		Runner:            ttySess.RunnerID,
		TerminalSessionID: termID,
		ListenAddr:        ttySess.Registry.ListenAddr,
	}

	enqueuedAt := time.Now()
	id, err := agentsend.EnqueueWith(home, sess, opts.prompt, agentsend.EnqueueOptions{NoSubmit: opts.noSubmit})
	if err != nil {
		return err
	}
	fmt.Println(id)

	waitOpts := agentsend.WaitOptions{
		EnqueuedAt:   enqueuedAt,
		Mode:         agentsend.WaitDefault,
		StartDrainer: true,
	}
	agentsend.StartDrainer(home, sess, provider)
	return agentsend.WaitForDelivery(home, sess, id, waitOpts)
}

func autoRunCreate(store agentstorage.Store, sessionID string, opts autoSendOrResumeOpts) error {
	if opts.detachFlag && opts.openFlag {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if opts.detachFlag && opts.jsonFlag {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if opts.prompt == "" && !opts.openFlag && !opts.detachFlag {
		return fmt.Errorf("prompt is required")
	}
	if opts.noSubmit && !opts.openFlag {
		return fmt.Errorf("--no-submit requires --open")
	}
	workspace, err := resolveRunDir(opts.dir)
	if err != nil {
		return err
	}
	runner := resolveCLIRunner(opts.agentRunner, opts.defaultRunner)
	if err := validateRunner(runner); err != nil {
		return err
	}
	if opts.openFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if opts.detachFlag && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--detach requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if err := requireTTYForSessionEnv(runner, opts.prependPaths, opts.envEntries); err != nil {
		return err
	}
	return agentui.Run(context.Background(), agentui.RunOptions{
		Prompt:                opts.prompt,
		Runner:                runner,
		Model:                 opts.model,
		SessionID:             sessionID,
		AgentRunnerBinary:     opts.agentRunnerBinary,
		AgentRunnerConfigHome: opts.agentRunnerConfigHome,
		PrependPaths:          opts.prependPaths,
		Env:                   opts.envEntries,
		JSON:                  opts.jsonFlag,
		Workspace:             workspace,
		KeepTerminalAlive:     opts.keepTTY || opts.openFlag || opts.detachFlag,
		Open:                  opts.openFlag,
		Detach:                opts.detachFlag,
		NoSubmit:              opts.noSubmit,
		Store:                 store,
		Stdout:                os.Stdout,
		Stderr:                os.Stderr,
	})
}

// tryResolveSessionMeta looks up a session by bare id or runner/session_id.
// found=false when the session is missing (MODE=run). Ambiguous ids error.
func tryResolveSessionMeta(store agentstorage.Store, ref string) (agentstorage.SessionMeta, bool, error) {
	meta, err := resolveSessionMeta(store, ref)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "session not found") {
			return agentstorage.SessionMeta{}, false, nil
		}
		return agentstorage.SessionMeta{}, false, err
	}
	return meta, true, nil
}

// resolveRunDir validates --dir when set: resolve relative paths against process
// cwd, require exists + directory, prefer EvalSymlinks. Empty means default
// (agentui uses Getwd).
func resolveRunDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("--dir: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("--dir: path does not exist: %s", abs)
		}
		return "", fmt.Errorf("--dir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("--dir: %s is not a directory", abs)
	}
	return abs, nil
}
