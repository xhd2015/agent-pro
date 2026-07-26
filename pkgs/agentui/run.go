package agentui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	codexagent "github.com/xhd2015/agent-pro/agent/cli/codex"
	crushagent "github.com/xhd2015/agent-pro/agent/cli/crush"
	grokagent "github.com/xhd2015/agent-pro/agent/cli/grok"
	piagent "github.com/xhd2015/agent-pro/agent/cli/pi"
	agentprovider "github.com/xhd2015/agent-pro/agent/cli/provider"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/event/codex_types"
	eventprint "github.com/xhd2015/agent-pro/agent/event/print"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// RunOptions configures a headless agent-run invocation.
type RunOptions struct {
	Prompt    string
	Runner    string
	Model     string
	SessionID string
	// TerminalSessionID is the TTY registry id. When empty, SessionID is used as
	// the custom terminal id (legacy --session / --session-id-from-prompt).
	// When PreferAutoTerminal is true, TerminalSessionID is ignored and the TTY
	// id is auto-allocated (session-N) while SessionID remains agent storage.
	TerminalSessionID string
	// PreferAutoTerminal forces auto session-N for the TTY registry even when
	// SessionID is set (used by resume fallback when zombie reclaim cannot free
	// the prior terminal id).
	PreferAutoTerminal    bool
	AgentRunnerBinary     string // optional binary name/path or "binary flag..." shell spec
	AgentRunnerConfigHome string // grok/codex data dir; falls back to AGENT_RUNNER_CONFIG_HOME
	// PrependPaths are absolute dirs prepended to the TTY child PATH (persisted on meta).
	PrependPaths []string
	// Env is ordered KEY=VALUE entries for the TTY child (persisted on meta; last-win).
	Env               []string
	JSON              bool
	Workspace         string
	Store             agentstorage.Store
	Stdout            io.Writer
	Stderr            io.Writer
	StreamPhases      bool // web: phased assistant start/update/end; CLI: single message events
	KeepTerminalAlive bool
	// Open is run --open: silent keep-alive TTY start, auto-attach, print id after detach.
	Open bool
	// Detach is run/resume --detach: silent keep-alive TTY daemon, soft grok bind,
	// print session-id + terminal-id on stdout, no attach / no event stream.
	Detach bool
	// NoSubmit is run --no-submit: with Open, inject prompt without trailing Enter.
	NoSubmit bool
	// Fork is run/resume --fork: pass grok --fork-session with ResumeSessionID (parent).
	// Requires grok-tty and a bound/resume parent Grok session id.
	Fork bool
	// ForkSessionID optional new Grok session UUID for --session-id with --fork-session.
	ForkSessionID string
	// Driver is the host re-exec config for TTY __serve_* children.
	Driver agentdriver.Driver
	// WebManagedGrokSync skips in-process grok sync; caller runs agentsync.EnsureGrokSync.
	WebManagedGrokSync bool
}

// Run executes the agent runner, streams output, and persists AgentEvents.
func Run(ctx context.Context, opts RunOptions) error {
	if opts.Store == nil {
		return fmt.Errorf("store is required")
	}
	if opts.Open || opts.Detach {
		opts.KeepTerminalAlive = true
	}
	// Empty prompt is allowed for --open / --detach / keep-alive reopen (resume
	// without followup). Headless one-shot still requires a prompt.
	if strings.TrimSpace(opts.Prompt) == "" && !opts.Open && !opts.Detach && !opts.KeepTerminalAlive {
		return fmt.Errorf("prompt is required")
	}
	runStart := time.Now()
	runner := strings.TrimSpace(opts.Runner)
	if runner == "" {
		cfg, err := opts.Store.Config()
		if err == nil && strings.TrimSpace(cfg.DefaultAgentRunner) != "" {
			runner = cfg.DefaultAgentRunner
		}
	}
	if runner == "" {
		runner = string(registry.AgentRunnerOpencode)
	}

	// userSessionID is set when --session or --session-id-from-prompt chose the id.
	// By default that id is also used as the TTY terminal registry id unless the
	// caller overrides via TerminalSessionID or PreferAutoTerminal (resume reclaim).
	userSessionID := strings.TrimSpace(opts.SessionID)
	sessionID := userSessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess_%d", os.Getpid())
	}
	ttySessionID := resolveTTYSessionID(opts, userSessionID)

	// Resolve workspace before session create so meta.workspace is always set
	// (explicit --dir / RunOptions.Workspace, or process cwd default).
	workspace := strings.TrimSpace(opts.Workspace)
	if workspace == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		workspace = wd
	}

	runnerSessionID := ""
	if sess, err := opts.Store.GetSession(sessionID); err == nil {
		runnerSessionID = strings.TrimSpace(sess.Meta.RunnerSessionID)
		_ = opts.Store.UpdateSessionStatus(sessionID, "running")
		// Persist session-scoped env injection (resume merge is done by caller).
		if len(opts.PrependPaths) > 0 || len(opts.Env) > 0 || strings.TrimSpace(opts.AgentRunnerConfigHome) != "" {
			_ = opts.Store.UpdateSessionEnvConfig(sessionID, opts.PrependPaths, opts.Env, opts.AgentRunnerConfigHome)
		}
	} else {
		createMeta := agentstorage.SessionMeta{
			Runner:                runner,
			SessionID:             sessionID,
			Status:                "running",
			Model:                 opts.Model,
			InitialPrompt:         strings.TrimSpace(opts.Prompt),
			Workspace:             workspace,
			PrependPaths:          append([]string(nil), opts.PrependPaths...),
			Env:                   append([]string(nil), opts.Env...),
			AgentRunnerConfigHome: strings.TrimSpace(opts.AgentRunnerConfigHome),
		}
		_ = opts.Store.CreateSession(sessionID, createMeta)
	}

	priorEvents, _, _ := opts.Store.ReadEvents(sessionID, 0)
	runnerPrompt := ResolveRunnerPrompt(runnerSessionID, opts.Prompt, priorEvents)

	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	env := agentexec.NewEnv(&agentexec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")

	appendEvent := func(ev types.AgentEvent) error {
		if err := opts.Store.AppendEvent(sessionID, ev); err != nil {
			return err
		}
		// --open / --detach stay silent: persist only, no human/JSON stream to the screen.
		if opts.Open || opts.Detach {
			return nil
		}
		line, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		if opts.JSON {
			_, err = fmt.Fprintln(stdout, string(line))
			return err
		}
		if formatted := eventprint.FormatAgentEventForStdout(ev); formatted != "" {
			_, err = fmt.Fprintln(stdout, formatted)
		}
		return err
	}
	emit := func(ev types.AgentEvent) error {
		if ev.Type == types.ActionMessage && strings.TrimSpace(ev.Role) == "" {
			ev.Role = "assistant"
		}
		if ev.Type == types.ActionMessage && ev.Timestamp == 0 {
			ev.Timestamp = time.Now().UnixMilli()
		}
		return appendEvent(ev)
	}

	persistTerminalSessionID := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		_ = opts.Store.UpdateSessionTerminalSessionID(sessionID, id)
	}
	webGrokManaged := runner == "grok-tty" && opts.WebManagedGrokSync
	// Open/Detach modes do not wait on discovery stream; skip in-process grok sync screen noise.
	grokSyncOwnsEvents := runner == "grok-tty" && opts.KeepTerminalAlive && !opts.WebManagedGrokSync && !opts.Open && !opts.Detach
	// Web-managed sync uses agentsync.EnsureGrokSync only; disable agenttty inline grok tail.
	// For Open/Detach, pass true so agenttty also skips inline discovery tails.
	ttyGrokSyncOwnsEvents := grokSyncOwnsEvents || webGrokManaged || opts.Open || opts.Detach
	if grokSyncOwnsEvents {
		startGrokSyncPoller(ctx, opts, emit)
		_ = ensureGrokSyncForSession(context.Background(), opts, resolveGrokSessionID(opts.Store, runner, sessionID), emit)
	}

	// Open + grok-tty: start bind worker early (before/alongside attach) so
	// discovery runs for the whole open lifetime and mid-open status can report
	// "binding". Detach uses a separate soft-only worker (1 minute, miss OK).
	// Use the original user prompt for matching (not continuation).
	var openBind *openGrokBindWorker
	if runner == "grok-tty" {
		bindPrompt := strings.TrimSpace(opts.Prompt)
		if bindPrompt == "" {
			bindPrompt = runnerPrompt
		}
		if opts.Open {
			openBind = startOpenGrokBindWorker(opts, runner, sessionID, workspace, bindPrompt, runStart, runnerSessionID)
		} else if opts.Detach {
			openBind = startDetachGrokBindWorker(opts, runner, sessionID, workspace, bindPrompt, runStart, runnerSessionID)
		}
	}

	newRunnerSessionID, newTerminalSessionID, runErr := streamRunner(ctx, runner, opts.Store.Home(), workspace, env, runnerPrompt, opts.Model, opts.AgentRunnerBinary, opts.AgentRunnerConfigHome, opts.PrependPaths, opts.Env, runnerSessionID, sessionID, ttySessionID, opts.StreamPhases, opts.KeepTerminalAlive, opts.Open, opts.Detach, opts.NoSubmit, opts.Fork, opts.ForkSessionID, ttyGrokSyncOwnsEvents, opts.Driver, persistTerminalSessionID, emit, stderr)
	if strings.TrimSpace(newRunnerSessionID) != "" {
		_ = opts.Store.UpdateSessionRunnerSessionID(sessionID, newRunnerSessionID)
	}
	if strings.TrimSpace(newTerminalSessionID) != "" {
		_ = opts.Store.UpdateSessionTerminalSessionID(sessionID, newTerminalSessionID)
	}
	if opts.Detach {
		if runErr != nil {
			if openBind != nil {
				openBind.Cancel()
				_ = openBind.WaitDone()
			}
			_ = opts.Store.UpdateSessionStatus(sessionID, "error")
			return runErr
		}
		// Soft bind: wait full detach budget (miss still exit 0). Status stays running.
		if openBind != nil {
			res := openBind.WaitDone()
			// Prefer streamRunner-discovered id if worker soft-missed.
			if strings.TrimSpace(res.id) == "" && strings.TrimSpace(newRunnerSessionID) != "" && res.err == nil {
				_ = opts.Store.UpdateSessionRunnerSessionID(sessionID, strings.TrimSpace(newRunnerSessionID))
				res.id = strings.TrimSpace(newRunnerSessionID)
			}
			// Soft only: ignore hard errors from unexpected paths; print lines on hit.
			_ = printOpenGrokBindResult(openGrokBindResult{
				id:          res.id,
				updatesPath: res.updatesPath,
				err:         nil, // detach never hard-fails on bind miss
				requireBind: false,
			}, stderr)
		}
		termID := strings.TrimSpace(newTerminalSessionID)
		if termID == "" {
			termID = strings.TrimSpace(ttySessionID)
		}
		_, _ = fmt.Fprintf(stdout, "session-id: %s\n", sessionID)
		_, _ = fmt.Fprintf(stdout, "terminal-id: %s\n", termID)
		// Leave meta.status = running (do not emit ActionDone / finished).
		return nil
	}
	if opts.Open {
		if runErr != nil {
			if openBind != nil {
				openBind.Cancel()
				_ = openBind.Wait()
			}
			_ = opts.Store.UpdateSessionStatus(sessionID, "error")
			return runErr
		}
		// After attach returns: print terminal id, then ALWAYS join bind worker.
		termID := strings.TrimSpace(newTerminalSessionID)
		if termID != "" {
			_, _ = fmt.Fprintf(stderr, "%s: %s\n", runner, termID)
		}
		if openBind != nil {
			// Prefer id already discovered mid-open; if streamRunner also
			// returned one and bind soft-missed, fall back to post-wait finalize
			// only when worker finished unbound without hard error (rare).
			res := openBind.Wait()
			if strings.TrimSpace(res.id) == "" && strings.TrimSpace(newRunnerSessionID) != "" && res.err == nil {
				// Worker may have started before known id appeared; re-run
				// finalize path only if hard-require still needs an id.
				if err := finalizeOpenGrokSession(ctx, opts, runner, sessionID, workspace, runnerPrompt, runStart, strings.TrimSpace(newRunnerSessionID), stderr); err != nil {
					_ = opts.Store.UpdateSessionStatus(sessionID, "error")
					return err
				}
				return nil
			}
			if err := printOpenGrokBindResult(res, stderr); err != nil {
				_ = opts.Store.UpdateSessionStatus(sessionID, "error")
				return err
			}
		} else if runner == "grok-tty" {
			// Non-worker path should not happen; keep finalize as safety net.
			if err := finalizeOpenGrokSession(ctx, opts, runner, sessionID, workspace, runnerPrompt, runStart, strings.TrimSpace(newRunnerSessionID), stderr); err != nil {
				_ = opts.Store.UpdateSessionStatus(sessionID, "error")
				return err
			}
		}
		return nil
	}
	if grokSyncOwnsEvents {
		if id := strings.TrimSpace(newRunnerSessionID); id != "" {
			_ = ensureGrokSyncForSession(context.Background(), opts, id, emit)
		}
	} else if !webGrokManaged {
		if err := emit(types.AgentEvent{Type: types.ActionDone}); err != nil {
			return err
		}
	}
	if grokSyncOwnsEvents || webGrokManaged {
		if runErr != nil {
			_ = opts.Store.UpdateSessionStatus(sessionID, "error")
		}
	} else {
		status := "finished"
		if runErr != nil {
			status = "error"
		}
		_ = opts.Store.UpdateSessionStatus(sessionID, status)
	}
	return runErr
}

// finalizeOpenGrokSession discovers the provider grok session after --open attach
// returns, persists runner_session_id, and prints the same stderr lines as non-open run.
//
// When discovery is "expected" (explicit GROK_HOME / config home / session-id hook),
// failure is a hard error. Otherwise discovery is best-effort so legacy --open
// lifecycle (fake TUI without a real grok home) still exits 0 after printing the
// terminal id.
func finalizeOpenGrokSession(ctx context.Context, opts RunOptions, runner, sessionID, workspace, prompt string, runStart time.Time, knownID string, stderr io.Writer) error {
	grokHome := agenttty.GrokHomeForRunner(opts.AgentRunnerConfigHome)
	requireBind := openGrokDiscoveryRequired(opts)
	id := strings.TrimSpace(knownID)
	updatesPath := ""
	if id != "" {
		if path, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, id); ok {
			updatesPath = path
		}
	}
	if id == "" || updatesPath == "" {
		// Discovery budget for post-attach finalize safety net.
		timeout := openGrokBindPostDetachGrace
		if !requireBind {
			timeout = openGrokBindSoftTimeout
		}
		discCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		// Prefer original user prompt for matching; fall back to continuation prompt.
		discoverPrompt := strings.TrimSpace(opts.Prompt)
		if discoverPrompt == "" {
			discoverPrompt = strings.TrimSpace(prompt)
		}
		discID, path, discErr := agenttty.DiscoverSession(discCtx, grokHome, workspace, discoverPrompt, runStart)
		if discErr != nil || strings.TrimSpace(discID) == "" {
			if openGrokHardFailOnUnresolved(opts) {
				return fmt.Errorf("error: grok session id not resolved for session %s", sessionID)
			}
			// Soft success: terminal id already printed by caller (attach-first).
			return nil
		}
		id = strings.TrimSpace(discID)
		updatesPath = path
	}
	if abs, absErr := filepath.Abs(updatesPath); absErr == nil {
		updatesPath = abs
	}
	_ = opts.Store.UpdateSessionRunnerSessionID(sessionID, id)
	_, _ = fmt.Fprintf(stderr, "grok-tty: grok session %s\n", id)
	_, _ = fmt.Fprintf(stderr, "grok-tty: grok updates %s\n", updatesPath)
	return nil
}

// openGrokDiscoveryRequired reports whether --open uses the hard discovery wait
// (full budget / post-detach grace) rather than the empty-prompt soft 750ms timeout.
//
// Non-empty prompt always hard-waits so real grok under default ~/.grok is not
// unbound after a soft race (see doc/LOOP_2026-07-11_open-bind-runner-unbound.md).
// Explicit config home / GROK_HOME / session-id hook also force hard wait.
// NoSubmit skips hard discovery: draft was never submitted, so no provider session
// is expected (soft unbound with short timeout).
// Whether an unresolved bind is a hard *error* is openGrokHardFailOnUnresolved.
func openGrokDiscoveryRequired(opts RunOptions) bool {
	if opts.NoSubmit {
		return false
	}
	if strings.TrimSpace(opts.Prompt) != "" {
		return true
	}
	return openGrokHardFailOnUnresolved(opts)
}

// openGrokHardFailOnUnresolved reports whether unresolved bind must fail the open.
//
// Explicit GROK_HOME / config home / session-id hook: hard error (status-resume
// isolation leaves). Non-empty prompt alone hard-waits discovery but still allows
// soft unbound on miss so attach-first --open (fake TUI / production path without
// an explicit grok home) can exit 0 after printing the terminal session id.
// NoSubmit always soft-unbounds: draft-only open never starts a provider turn.
// Detach always soft-unbounds: miss exits 0 after printing both ids.
func openGrokHardFailOnUnresolved(opts RunOptions) bool {
	if opts.Detach || opts.NoSubmit {
		return false
	}
	if strings.TrimSpace(opts.AgentRunnerConfigHome) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("GROK_HOME")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("AGENT_RUNNER_CONFIG_HOME")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("AGENT_RUN_GROK_TTY_GROK_SESSION_ID")) != "" {
		return true
	}
	return false
}

// resolveTTYSessionID picks the custom terminal registry id for streamRunner.
// PreferAutoTerminal → empty (auto session-N). Explicit TerminalSessionID wins.
// Otherwise fall back to userSessionID (legacy --session-id behavior).
func resolveTTYSessionID(opts RunOptions, userSessionID string) string {
	if opts.PreferAutoTerminal {
		return ""
	}
	if id := strings.TrimSpace(opts.TerminalSessionID); id != "" {
		return id
	}
	return userSessionID
}

// streamRunner runs the selected agent. ttySessionID is the custom terminal
// registry id (from --session / --session-id-from-prompt); empty keeps session-N.
func streamRunner(ctx context.Context, runner, home, workspace string, env *agentexec.Env, prompt, model, agentRunnerBinary, agentRunnerConfigHome string, prependPaths, envEntries []string, runnerSessionID, agentSessionID, ttySessionID string, streamPhases, keepTerminalAlive, open, detach, noSubmit, fork bool, forkSessionID string, grokSyncOwnsEvents bool, driver agentdriver.Driver, onTerminalSessionID func(string), emit func(types.AgentEvent) error, stderr io.Writer) (string, string, error) {
	if agenttty.IsTTYRunner(runner) {
		terminalSessionID := ""
		onID := func(id string) {
			terminalSessionID = strings.TrimSpace(id)
			if onTerminalSessionID != nil {
				onTerminalSessionID(terminalSessionID)
			}
		}
		newRunnerSessionID, _, err := agenttty.RunHeadless(ctx, agenttty.RunOptions{
			Home:                  home,
			Workspace:             workspace,
			Prompt:                prompt,
			Model:                 model,
			ResumeSessionID:       runnerSessionID,
			Fork:                  fork,
			ForkSessionID:         forkSessionID,
			RunnerID:              runner,
			SessionID:             ttySessionID,
			AgentSessionID:        agentSessionID,
			AgentPath:             agentRunnerBinary,
			AgentRunnerConfigHome: agentRunnerConfigHome,
			PrependPaths:          prependPaths,
			Env:                   envEntries,
			Driver:                driver,
			KeepTerminalAlive:     keepTerminalAlive || open || detach,
			Open:                  open,
			Detach:                detach,
			NoSubmit:              noSubmit,
			GrokSyncOwnsEvents:    grokSyncOwnsEvents,
			Stderr:                stderr,
			Emit:                  emit,
			OnTerminalSessionID:   onID,
		})
		if strings.TrimSpace(newRunnerSessionID) != "" {
			return newRunnerSessionID, terminalSessionID, err
		}
		return runnerSessionID, terminalSessionID, err
	}
	switch registry.AgentRunnerID(runner) {
	case registry.AgentRunnerFakeCodex, registry.AgentRunnerCodex:
		err := streamCodexLike(ctx, runner, workspace, env, prompt, model, streamPhases, emit, stderr)
		return "", "", err
	default:
		if _, err := agentprovider.Build(registry.AgentRunnerID(runner), "", workspace, env); err != nil {
			return "", "", err
		}
		runnerSessionID, err := streamViaRegistry(ctx, runner, workspace, env, prompt, model, runnerSessionID, streamPhases, emit)
		return runnerSessionID, "", err
	}
}

func streamViaRegistry(ctx context.Context, runner, workspace string, env *agentexec.Env, prompt, model, runnerSessionID string, streamPhases bool, emit func(types.AgentEvent) error) (string, error) {
	built, err := agentprovider.Build(registry.AgentRunnerID(runner), "", workspace, env)
	if err != nil {
		return "", err
	}

	streamID := newAssistantStreamID()
	var buf strings.Builder
	started := false

	_, err = built.Agent.Ask(ctx, prompt, &registry.AskOptions{
		Model:     model,
		Workspace: workspace,
		SessionID: runnerSessionID,
	}, func(delta string) {
		if delta == "" {
			return
		}
		buf.WriteString(delta)
		if !streamPhases {
			return
		}
		if !started {
			_ = emit(types.AgentEvent{
				ID:        streamID,
				Type:      types.ActionMessage,
				Role:      "assistant",
				Phase:     types.PhaseStart,
				Timestamp: time.Now().UnixMilli(),
			})
			started = true
		}
		_ = emit(types.AgentEvent{
			ID:        streamID,
			Type:      types.ActionMessage,
			Role:      "assistant",
			Phase:     types.PhaseUpdate,
			Text:      buf.String(),
			Timestamp: time.Now().UnixMilli(),
		})
	})
	if err != nil {
		return runnerSessionIDFromAgent(built.Agent), err
	}

	text := strings.TrimSpace(buf.String())
	if text == "" {
		return runnerSessionIDFromAgent(built.Agent), nil
	}
	if streamPhases {
		if started {
			if err := emit(types.AgentEvent{
				ID:        streamID,
				Type:      types.ActionMessage,
				Role:      "assistant",
				Phase:     types.PhaseEnd,
				Text:      text,
				Timestamp: time.Now().UnixMilli(),
			}); err != nil {
				return runnerSessionIDFromAgent(built.Agent), err
			}
		} else if err := emitPhasedAssistantMessage(emit, streamID, text); err != nil {
			return runnerSessionIDFromAgent(built.Agent), err
		}
	} else {
		if err := emit(types.AgentEvent{
			Type:      types.ActionMessage,
			Role:      "assistant",
			Text:      text,
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			return runnerSessionIDFromAgent(built.Agent), err
		}
	}

	return runnerSessionIDFromAgent(built.Agent), nil
}

func runnerSessionIDFromAgent(agent registry.Agent) string {
	switch a := agent.(type) {
	case *grokagent.GrokAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *piagent.PiAgent:
		return strings.TrimSpace(a.LastSessionID)
	case *crushagent.CrushAgent:
		return strings.TrimSpace(a.LastSessionID)
	}
	return ""
}

func streamCodexLike(ctx context.Context, runner, workspace string, env *agentexec.Env, prompt, model string, streamPhases bool, emit func(types.AgentEvent) error, stderr io.Writer) error {
	built, err := agentprovider.Build(registry.AgentRunnerID(runner), "", workspace, env)
	if err != nil {
		return err
	}
	agentPath := ""
	if ca, ok := built.Agent.(*codexagent.CodexAgent); ok && ca.AgentPath != "" {
		agentPath = ca.AgentPath
	}
	if agentPath == "" {
		var lookErr error
		agentPath, lookErr = env.LookPath(runner)
		if lookErr != nil {
			return lookErr
		}
	}

	args := []string{"exec", "--json", "--skip-git-repo-check", "--cd", workspace}
	if model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, prompt)

	cmd := env.CommandContext(ctx, agentPath, args...)
	cmd.Dir = workspace
	cmd.Stderr = stderr
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	reader := bufio.NewReader(pipe)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSpace(line)
			if line != "" && strings.HasPrefix(line, "{") {
				var codexEv codex_types.Event
				if err := json.Unmarshal([]byte(line), &codexEv); err == nil {
					for _, ev := range codex_types.FromCodex([]codex_types.Event{codexEv}, "") {
						if ev.Type == types.ActionMessage {
							if strings.TrimSpace(ev.Role) == "" {
								ev.Role = "assistant"
							}
							if ev.Role == "assistant" && ev.Phase == "" {
								if streamPhases {
									if err := emitPhasedAssistantMessage(emit, "", ev.Text); err != nil {
										return err
									}
								} else if err := emit(ev); err != nil {
									return err
								}
								continue
							}
						}
						if err := emit(ev); err != nil {
							return err
						}
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return cmd.Wait()
}
