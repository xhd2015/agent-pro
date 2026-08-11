package agentrunapi

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
)

// defaultSendLive enqueues a follow-up on a live TTY session (agentsend parity).
func defaultSendLive(ctx context.Context, opts Opts, meta agentstorage.SessionMeta) error {
	_ = ctx
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	// --open/--detach/--new-terminal are accepted but ignored while live.
	if opts.Open {
		fmt.Fprintln(stderr, "note: --open ignored while session is live; sending follow-up")
	}
	if opts.Detach {
		fmt.Fprintln(stderr, "note: --detach ignored while session is live; sending follow-up")
	}
	if opts.NewTerminal {
		fmt.Fprintln(stderr, "note: --new-terminal ignored while session is live; sending follow-up")
	}
	if strings.TrimSpace(opts.Prompt) == "" {
		fmt.Fprintln(stderr, "warning: session is live; no message to send")
		return nil
	}

	termID := strings.TrimSpace(meta.TerminalSessionID)
	if termID == "" {
		termID = strings.TrimSpace(meta.SessionID)
	}
	if termID == "" {
		return fmt.Errorf("cannot send: missing terminal_session_id and session_id")
	}

	home := opts.Store.Home()
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
	id, err := agentsend.EnqueueWith(home, sess, opts.Prompt, agentsend.EnqueueOptions{NoSubmit: opts.NoSubmit})
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, id)

	waitOpts := agentsend.WaitOptions{
		EnqueuedAt:   enqueuedAt,
		Mode:         agentsend.WaitDefault,
		StartDrainer: true,
	}
	agentsend.StartDrainer(home, sess, provider)
	return agentsend.WaitForDelivery(home, sess, id, waitOpts)
}

// defaultRunSession creates/runs a session via agentui (MODE=run).
func defaultRunSession(ctx context.Context, opts Opts, meta agentstorage.SessionMeta, found bool) error {
	_ = meta
	_ = found
	if opts.Detach && opts.Open {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if opts.Detach && opts.JSON {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if opts.Open && opts.JSON {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if strings.TrimSpace(opts.Prompt) == "" && !opts.Open && !opts.Detach {
		return fmt.Errorf("prompt is required")
	}
	if opts.NoSubmit && !opts.Open {
		return fmt.Errorf("--no-submit requires --open")
	}
	workspace, err := resolveWorkspaceDir(opts.WorkspaceDir)
	if err != nil {
		return err
	}
	runner := strings.TrimSpace(opts.AgentRunner)
	if runner == "" {
		runner = "grok-tty"
	}
	if opts.Open && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if opts.Detach && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--detach requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}

	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	return agentui.Run(ctx, agentui.RunOptions{
		Prompt:                opts.Prompt,
		Runner:                runner,
		Model:                 opts.Model,
		ModelReasoningEffort:  opts.ModelReasoningEffort,
		SessionID:             opts.SessionID,
		AgentRunnerBinary:     opts.AgentRunnerBinary,
		AgentRunnerConfigHome: opts.RunnerConfigHome,
		PrependPaths:          opts.PrependPaths,
		Env:                   opts.Env,
		Color:                 opts.Color,
		JSON:                  opts.JSON,
		Workspace:             workspace,
		KeepTerminalAlive: keepAliveForOpenOrDetach(opts.KeepTTY, opts.Open, opts.Detach),
		Open:              opts.Open,
		Detach:            opts.Detach,
		NoSubmit:          opts.NoSubmit,
		Driver:            opts.Driver,
		Store:             opts.Store,
		Stdout:            stdout,
		Stderr:            stderr,
	})
}

// defaultResumeSession reopens a bound+exited session via agentui.
// Full CLI resume (reclaim, grok relocate, merge meta env) is typically injected
// by cmd/agent-run; this default covers library callers with basic resume.
func defaultResumeSession(ctx context.Context, opts Opts, meta agentstorage.SessionMeta) error {
	prompt := strings.TrimSpace(opts.Prompt)
	keepTTY := opts.KeepTTY
	if prompt == "" && !opts.Open && !opts.Detach {
		keepTTY = true
	}
	if opts.Detach && opts.Open {
		return fmt.Errorf("--detach and --open are mutually exclusive; cannot use both")
	}
	if opts.Detach && opts.JSON {
		return fmt.Errorf("--detach and --json are mutually exclusive; cannot use both")
	}
	if opts.Open && opts.JSON {
		return fmt.Errorf("--open and --json are mutually exclusive; cannot use both")
	}
	if opts.NoSubmit && !opts.Open {
		return fmt.Errorf("--no-submit requires --open")
	}
	if strings.TrimSpace(meta.RunnerSessionID) == "" {
		return fmt.Errorf("runner session not bound (missing runner_session_id); cannot resume")
	}

	runner := strings.TrimSpace(opts.AgentRunner)
	if runner == "" {
		runner = strings.TrimSpace(meta.Runner)
	}
	if runner == "" {
		runner = "grok-tty"
	}
	if opts.Open && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--open requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}
	if opts.Detach && !agenttty.IsTTYRunner(runner) {
		return fmt.Errorf("--detach requires a TTY runner (got %s); non-TTY runners like fake-codex are not supported", runner)
	}

	workspace, err := resolveWorkspaceDir(opts.WorkspaceDir)
	if err != nil {
		return err
	}
	if workspace == "" {
		if w := strings.TrimSpace(meta.Workspace); w != "" {
			workspace = w
		}
	}
	model := strings.TrimSpace(opts.Model)
	if model == "" {
		model = strings.TrimSpace(meta.Model)
	}
	configHome := strings.TrimSpace(opts.RunnerConfigHome)
	if configHome == "" {
		configHome = strings.TrimSpace(meta.AgentRunnerConfigHome)
	}

	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	sessionID := strings.TrimSpace(meta.SessionID)
	if sessionID == "" {
		sessionID = opts.SessionID
	}

	return agentui.Run(ctx, agentui.RunOptions{
		Prompt:                prompt,
		Runner:                runner,
		Model:                 model,
		ModelReasoningEffort:  opts.ModelReasoningEffort,
		SessionID:             sessionID,
		AgentRunnerBinary:     opts.AgentRunnerBinary,
		AgentRunnerConfigHome: configHome,
		PrependPaths:          appendStringLists(meta.PrependPaths, opts.PrependPaths),
		Env:                   appendStringLists(meta.Env, opts.Env),
		Color:                 opts.Color,
		JSON:                  opts.JSON,
		Workspace:             workspace,
		KeepTerminalAlive: keepAliveForOpenOrDetach(keepTTY, opts.Open, opts.Detach),
		Open:              opts.Open,
		Detach:            opts.Detach,
		NoSubmit:          opts.NoSubmit,
		Driver:            opts.Driver,
		Store:             opts.Store,
		Stdout:            stdout,
		Stderr:            stderr,
	})
}

// keepAliveForOpenOrDetach mirrors agentui: detach always keeps; open keeps
// unless AGENT_RUN_OPEN_CLOSE_EXITS experiment is on.
func keepAliveForOpenOrDetach(keepTTY, open, detach bool) bool {
	if keepTTY || detach {
		return true
	}
	if open && !agenttty.OpenCloseExitsExperiment() {
		return true
	}
	return false
}

func resolveWorkspaceDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("workspace dir: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("workspace dir: path does not exist: %s", abs)
		}
		return "", fmt.Errorf("workspace dir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace dir: %s is not a directory", abs)
	}
	return abs, nil
}

func appendStringLists(base, extra []string) []string {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}
	out := make([]string, 0, len(base)+len(extra))
	out = append(out, base...)
	out = append(out, extra...)
	return out
}
