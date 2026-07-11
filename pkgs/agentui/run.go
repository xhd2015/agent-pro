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
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// RunOptions configures a headless agent-run invocation.
type RunOptions struct {
	Prompt              string
	Runner              string
	Model               string
	SessionID           string
	AgentRunnerBinary     string // optional binary name/path or "binary flag..." shell spec
	AgentRunnerConfigHome string // grok/codex data dir; falls back to AGENT_RUNNER_CONFIG_HOME
	JSON                bool
	Workspace           string
	Store               agentstorage.Store
	Stdout              io.Writer
	Stderr              io.Writer
	StreamPhases        bool // web: phased assistant start/update/end; CLI: single message events
	KeepTerminalAlive   bool
	// Open is run --open: silent keep-alive TTY start, auto-attach, print id after detach.
	Open bool
	// NoSubmit is run --no-submit: with Open, inject prompt without trailing Enter.
	NoSubmit bool
	// WebManagedGrokSync skips in-process grok sync; caller runs agentsync.EnsureGrokSync.
	WebManagedGrokSync bool
}

// Run executes the agent runner, streams output, and persists AgentEvents.
func Run(ctx context.Context, opts RunOptions) error {
	if opts.Store == nil {
		return fmt.Errorf("store is required")
	}
	if opts.Open {
		opts.KeepTerminalAlive = true
	}
	if strings.TrimSpace(opts.Prompt) == "" && !opts.Open {
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
	// Only then do we also use it as the TTY terminal registry id.
	userSessionID := strings.TrimSpace(opts.SessionID)
	sessionID := userSessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess_%d", os.Getpid())
	}

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
	if sess, err := opts.Store.GetSession(runner, sessionID); err == nil {
		runnerSessionID = strings.TrimSpace(sess.Meta.RunnerSessionID)
		_ = opts.Store.UpdateSessionStatus(runner, sessionID, "running")
	} else {
		createMeta := agentstorage.SessionMeta{
			Runner:        runner,
			SessionID:     sessionID,
			Status:        "running",
			Model:         opts.Model,
			InitialPrompt: strings.TrimSpace(opts.Prompt),
			Workspace:     workspace,
		}
		_ = opts.Store.CreateSession(runner, sessionID, createMeta)
	}

	priorEvents, _, _ := opts.Store.ReadEvents(runner, sessionID, 0)
	runnerPrompt := BuildContinuationPrompt(priorEvents, opts.Prompt)

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
		if err := opts.Store.AppendEvent(runner, sessionID, ev); err != nil {
			return err
		}
		// --open stays silent: persist only, no human/JSON stream to the screen.
		if opts.Open {
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
		_ = opts.Store.UpdateSessionTerminalSessionID(runner, sessionID, id)
	}
	webGrokManaged := runner == "grok-tty" && opts.WebManagedGrokSync
	// Open mode does not wait on discovery; skip in-process grok sync screen noise.
	grokSyncOwnsEvents := runner == "grok-tty" && opts.KeepTerminalAlive && !opts.WebManagedGrokSync && !opts.Open
	// Web-managed sync uses agentsync.EnsureGrokSync only; disable agenttty inline grok tail.
	// For Open, pass true so agenttty also skips inline discovery tails.
	ttyGrokSyncOwnsEvents := grokSyncOwnsEvents || webGrokManaged || opts.Open
	if grokSyncOwnsEvents {
		startGrokSyncPoller(ctx, opts, emit)
		_ = ensureGrokSyncForSession(context.Background(), opts, resolveGrokSessionID(opts.Store, runner, sessionID), emit)
	}
	newRunnerSessionID, newTerminalSessionID, runErr := streamRunner(ctx, runner, opts.Store.Home(), workspace, env, runnerPrompt, opts.Model, opts.AgentRunnerBinary, opts.AgentRunnerConfigHome, runnerSessionID, sessionID, userSessionID, opts.StreamPhases, opts.KeepTerminalAlive, opts.Open, opts.NoSubmit, ttyGrokSyncOwnsEvents, persistTerminalSessionID, emit, stderr)
	if strings.TrimSpace(newRunnerSessionID) != "" {
		_ = opts.Store.UpdateSessionRunnerSessionID(runner, sessionID, newRunnerSessionID)
	}
	if strings.TrimSpace(newTerminalSessionID) != "" {
		_ = opts.Store.UpdateSessionTerminalSessionID(runner, sessionID, newTerminalSessionID)
	}
	if opts.Open {
		if runErr != nil {
			_ = opts.Store.UpdateSessionStatus(runner, sessionID, "error")
			return runErr
		}
		// After attach returns: print terminal id, then finalize grok session bind.
		termID := strings.TrimSpace(newTerminalSessionID)
		if termID != "" {
			_, _ = fmt.Fprintf(stderr, "%s: %s\n", runner, termID)
		}
		if runner == "grok-tty" {
			if err := finalizeOpenGrokSession(ctx, opts, runner, sessionID, workspace, runnerPrompt, runStart, strings.TrimSpace(newRunnerSessionID), stderr); err != nil {
				_ = opts.Store.UpdateSessionStatus(runner, sessionID, "error")
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
			_ = opts.Store.UpdateSessionStatus(runner, sessionID, "error")
		}
	} else {
		status := "finished"
		if runErr != nil {
			status = "error"
		}
		_ = opts.Store.UpdateSessionStatus(runner, sessionID, status)
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
		// Short grace for on-disk session materialization (or env session-id hook).
		timeout := 5 * time.Second
		if !requireBind {
			// Best-effort: fail fast when no config home / hook is configured.
			timeout = 750 * time.Millisecond
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
			if requireBind {
				return fmt.Errorf("error: grok session id not resolved for session %s", sessionID)
			}
			// Soft success: terminal id already printed by caller.
			return nil
		}
		id = strings.TrimSpace(discID)
		updatesPath = path
	}
	if abs, absErr := filepath.Abs(updatesPath); absErr == nil {
		updatesPath = abs
	}
	_ = opts.Store.UpdateSessionRunnerSessionID(runner, sessionID, id)
	_, _ = fmt.Fprintf(stderr, "grok-tty: grok session %s\n", id)
	_, _ = fmt.Fprintf(stderr, "grok-tty: grok updates %s\n", updatesPath)
	return nil
}

// openGrokDiscoveryRequired reports whether --open must resolve a grok session id.
func openGrokDiscoveryRequired(opts RunOptions) bool {
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

// streamRunner runs the selected agent. ttySessionID is the custom terminal
// registry id (from --session / --session-id-from-prompt); empty keeps session-N.
func streamRunner(ctx context.Context, runner, home, workspace string, env *agentexec.Env, prompt, model, agentRunnerBinary, agentRunnerConfigHome, runnerSessionID, agentSessionID, ttySessionID string, streamPhases, keepTerminalAlive, open, noSubmit, grokSyncOwnsEvents bool, onTerminalSessionID func(string), emit func(types.AgentEvent) error, stderr io.Writer) (string, string, error) {
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
			RunnerID:              runner,
			SessionID:             ttySessionID,
			AgentSessionID:        agentSessionID,
			AgentPath:             agentRunnerBinary,
			AgentRunnerConfigHome: agentRunnerConfigHome,
			KeepTerminalAlive:     keepTerminalAlive || open,
			Open:                  open,
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
