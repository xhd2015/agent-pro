package agenttty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// RunOptions configures a headless TTY runner invocation via detached serve.
type RunOptions struct {
	Home      string
	Workspace string
	Prompt    string
	Model     string
	// ModelReasoningEffort is optional Codex -c model_reasoning_effort=<level>.
	// Applied after BuildArgv for codex-tty only; empty leaves Codex home config.
	ModelReasoningEffort string
	ResumeSessionID      string
	// Fork appends grok --fork-session after --resume (grok-tty only). Requires
	// non-empty ResumeSessionID. Creates a new Grok conversation branch.
	Fork bool
	// ForkSessionID optional Grok --session-id for the forked session (with Fork).
	ForkSessionID string
	RunnerID      string
	// SessionID is the terminal registry id. Empty → auto-reserve session-N.
	// When set (explicit --session or --session-id-from-prompt), storage and registry share it.
	SessionID             string
	AgentSessionID        string
	SettingsPath          string
	AgentPath             string
	AgentRunnerConfigHome string
	// PrependPaths are absolute dirs prepended to the PTY child PATH (ordered).
	PrependPaths []string
	// Env is ordered KEY=VALUE entries applied to the PTY child (last-win per key).
	Env []string
	// Color forces TTY child color env last (unset NO_COLOR; FORCE_COLOR/CLICOLOR;
	// TERM fixup when empty/dumb). Not persisted on session meta.
	Color bool
	// ExitOnIdle / IdleTimeout are launch-time idle-exit flags (not persisted
	// on session meta). When enabled, RunHeadless writes idle-policy.json
	// before ttywatch.HeadlessRun.
	ExitOnIdle  bool
	IdleTimeout time.Duration
	// Driver is the host re-exec config for __serve_* children (see agentdriver).
	// Zero → DefaultSelf. Prefer over BinaryPath.
	Driver agentdriver.Driver
	// BinaryPath is deprecated: use Driver.Binary.
	BinaryPath        string
	KeepTerminalAlive bool
	// Open is run --open: silent start, optional inject, auto-attach, no pre-attach id print.
	Open bool
	// Detach is run/resume --detach: keep-alive daemon, no attach, no stream tail;
	// caller prints session-id/terminal-id on stdout after return.
	Detach bool
	// NoSubmit injects the prompt without trailing Enter (suffixCR=false). Used with Open.
	NoSubmit            bool
	GrokSyncOwnsEvents  bool
	Stderr              io.Writer
	Emit                func(types.AgentEvent) error
	OnTerminalSessionID func(string)
	// OnRunnerSessionID is invoked as soon as a provider session id is bound
	// (mid-open), so status can report bound before attach returns.
	OnRunnerSessionID func(string)
}

// recordSessionError is intentionally best-effort: runtime diagnostics must
// never be sent into an attached provider TTY, or turn a secondary log-write
// failure into a failed agent run.
func recordSessionError(opts RunOptions, component string, err error) {
	_ = agentstorage.AppendErrorLog(opts.Home, opts.AgentSessionID, component, err)
}

// RunHeadless starts a detached ttywatch serve session, injects the prompt, tails
// agent events, and waits for the child to exit unless KeepTerminalAlive is set.
// When Open is set, starts keep-alive, injects only if prompt non-empty, auto-attaches,
// and returns after attach without printing the session id (caller prints after).
// When Detach is set, starts keep-alive, optional inject for resume follow-up, and
// returns immediately without attach or event stream (caller prints both ids).
func RunHeadless(ctx context.Context, opts RunOptions) (runnerSessionID, terminalSessionID string, err error) {
	ensureStubRegistered()
	runnerID := strings.TrimSpace(opts.RunnerID)
	if runnerID == "" {
		runnerID = "grok-tty"
	}
	provider, ok := Get(runnerID)
	if !ok {
		return "", "", fmt.Errorf("unknown TTY runner: %s", runnerID)
	}
	// Detach: keep-alive daemon is the product.
	// Open: by default OpenCloseExits() — do NOT force keep-alive so window
	// close (writer disconnect → stopChild) also tears down __serve__.
	// Opt out: AGENT_RUN_OPEN_CLOSE_EXITS=0 restores pre-fix keep-alive open.
	// Explicit KeepTerminalAlive / --keep-tty still wins when already set.
	if opts.Detach {
		opts.KeepTerminalAlive = true
	} else if opts.Open {
		if !OpenCloseExits() {
			opts.KeepTerminalAlive = true
		}
	}
	// commandcode-tty: always keep PTY alive for snapshot/attach after run/open.
	// OpenCloseExits still tears down on window-close via attach_mode=open; the
	// serve must remain reachable for post-open snapshot probes in the meantime.
	if runnerID == "commandcode-tty" {
		opts.KeepTerminalAlive = true
	}
	// codex-tty headless: keep serve alive through inject+turn. Without keep-alive,
	// /input can 404 mid-inject on short-lived serve teardown ("inject endpoint not found").
	if runnerID == "codex-tty" && !opts.Open && !opts.Detach {
		opts.KeepTerminalAlive = true
	}
	// Empty prompt OK for open / detach / keep-alive reopen (e.g. resume without followup).
	if strings.TrimSpace(opts.Prompt) == "" && !opts.Open && !opts.Detach && !opts.KeepTerminalAlive {
		return "", "", fmt.Errorf("prompt is required")
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Workspace == "" {
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			return "", "", wdErr
		}
		opts.Workspace = wd
	}

	cfg := runConfig{
		runnerID:       runnerID,
		bannerProvider: provider.BannerProvider,
		bannerMarkers:  provider.BannerMarkers,
	}

	env := newExecEnv()
	argv, err := provider.BuildArgv(env, opts.SettingsPath, opts.AgentPath, opts.Model, opts.ResumeSessionID)
	if err != nil {
		return "", "", err
	}
	// codex-tty: optional model_reasoning_effort after BuildArgv (BuildArgvFunc unchanged).
	if runnerID == "codex-tty" {
		argv = ApplyCodexReasoningEffort(argv, opts.ModelReasoningEffort)
	}
	if opts.Fork {
		if runnerID != "grok-tty" {
			return "", "", fmt.Errorf("--fork requires grok-tty (got %s)", runnerID)
		}
		if strings.TrimSpace(opts.ResumeSessionID) == "" {
			return "", "", fmt.Errorf("--fork requires a resume parent session id")
		}
		if !hasFlagPair(argv, "--fork-session") {
			argv = append(argv, "--fork-session")
		}
		if sid := strings.TrimSpace(opts.ForkSessionID); sid != "" && !hasFlagPair(argv, "--session-id") {
			argv = append(argv, "--session-id", sid)
		}
	}
	// New session (no --resume): pass initial prompt as trailing positional arg
	// (grok [PROMPT]). Resume follow-ups stay inject-only so argv keeps --resume.
	// See appendNewSessionPrompt for runner-specific rules and UTF-8 contract.
	if strings.TrimSpace(opts.ResumeSessionID) == "" {
		argv = appendNewSessionPrompt(argv, runnerID, opts.Prompt, opts.NoSubmit, opts.Open)
	}

	configHome := ResolveAgentRunnerConfigHome(opts.AgentRunnerConfigHome)
	if configHome == "" {
		if provisioned, provErr := AutoProvisionGrokConfigHome(runnerID, opts.AgentPath, ""); provErr != nil {
			return "", "", provErr
		} else if provisioned != "" {
			configHome = provisioned
		}
	}
	envSpec := BuildChildProcessEnv(runnerID, configHome, opts.PrependPaths, opts.Env, opts.Color, os.Getenv("TERM"))

	driver := opts.Driver
	if strings.TrimSpace(driver.Binary) == "" && strings.TrimSpace(opts.BinaryPath) != "" {
		driver.Binary = opts.BinaryPath
	}

	if err := writeIdlePolicyBeforeServe(opts); err != nil {
		return "", "", err
	}

	result, err := ttywatch.HeadlessRun(ctx, ttywatch.HeadlessRunOptions{
		Home:           opts.Home,
		RegistrySubdir: provider.RegistryDir,
		SessionID:      strings.TrimSpace(opts.SessionID),
		Driver:         driver,
		Command:        argv,
		CommandEnv:     envSpec.Set,
		CommandUnset:   envSpec.Unset,
		Cwd:            opts.Workspace,
		KeepAlive:      opts.KeepTerminalAlive,
	})
	if err != nil {
		return "", "", err
	}

	sessionID := result.SessionID
	listenAddr := result.Entry.ListenAddr
	terminalSessionID = sessionID
	if opts.OnTerminalSessionID != nil {
		opts.OnTerminalSessionID(sessionID)
	}
	if opts.AgentSessionID != "" {
		go dualWriteAfterRegistry(opts.Home, runnerID, opts.AgentSessionID, sessionID, provider)
	}

	// Open/Detach: caller prints ids (open after attach; detach immediately on stdout).
	if !opts.Open && !opts.Detach {
		fmt.Fprintf(opts.Stderr, "%s: %s\n", runnerID, sessionID)
	}

	runStart := time.Now()
	// Normalize before inject paths (resume / NoSubmit / codex). New-session argv
	// also normalizes inside appendNewSessionPrompt.
	promptText := prepareRunnerPrompt(runnerID, strings.TrimSpace(opts.Prompt))
	isResume := strings.TrimSpace(opts.ResumeSessionID) != ""

	// --detach: keep-alive daemon only — no attach, no event stream, no wait for turn.
	// Prompt on argv for new sessions when non-empty and not NoSubmit (same as non-open).
	// Resume follow-up injects after soft banner wait; empty prompt is reopen-only.
	if opts.Detach {
		if promptText != "" && (isResume || opts.NoSubmit || runnerID == "codex-tty") {
			if readyErr := waitForBannerRemote(ctx, listenAddr, sessionID, provider.BannerProvider, provider.BannerMarkers); readyErr != nil {
				if ctx.Err() != nil {
					return "", terminalSessionID, ctx.Err()
				}
				// Soft: skip inject on banner timeout; still leave daemon alive.
			} else if err := InjectMessage(listenAddr, sessionID, runnerID, promptText, !opts.NoSubmit); err != nil {
				return "", terminalSessionID, err
			}
		} else {
			// Brief soft ready so registry/PTY is fully up before parent exits.
			_ = waitForOpenReady(ctx, listenAddr, sessionID, 3*time.Second)
			if isCodexProvider(provider.BannerProvider) || runnerID == "codex-tty" {
				acceptCodexTrustRemote(ctx, listenAddr, sessionID, provider.BannerProvider, 15*time.Second, nil)
			}
			select {
			case <-ctx.Done():
				return "", terminalSessionID, ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}

		if terminalSessionID != "" && opts.AgentSessionID != "" && result.Entry != nil {
			_ = WriteTTYJSON(opts.Home, TTYSnapshot{
				RunnerID:          runnerID,
				AgentSessionID:    opts.AgentSessionID,
				TerminalSessionID: terminalSessionID,
				ListenAddr:        result.Entry.ListenAddr,
				PID:               result.Entry.PID,
				CreatedAt:         result.Entry.CreatedAt,
				ScreenStatus:      "unknown",
				Alive:             true,
			})
		}
		// Leave PTY/registry alive for later attach/send; bind codex id when possible.
		return resolveOpenDetachRunnerSessionID(ctx, opts, runnerID, provider, configHome, runStart, listenAddr, sessionID), terminalSessionID, nil
	}

	// --open: attach as soon as the PTY is registered. Trust, prompt inject, and
	// Codex session bind continue in the background so bootstrap stays visible.
	// AGENT_RUN_OPEN_ATTACH_INSTANT=1 skips interactive attach (tests).
	if opts.Open {
		agentGone := func() bool {
			return RegistryAgentExited(opts.Home, runnerID, sessionID)
		}
		if terminalSessionID != "" && opts.AgentSessionID != "" && result.Entry != nil {
			_ = WriteTTYJSON(opts.Home, TTYSnapshot{
				RunnerID:          runnerID,
				AgentSessionID:    opts.AgentSessionID,
				TerminalSessionID: terminalSessionID,
				ListenAddr:        result.Entry.ListenAddr,
				PID:               result.Entry.PID,
				CreatedAt:         result.Entry.CreatedAt,
				ScreenStatus:      "unknown",
				Alive:             true,
			})
		}
		attachMode := "attach"
		if OpenCloseExits() {
			attachMode = "open"
		}

		codexOpen := isCodexProvider(provider.BannerProvider) || runnerID == "codex-tty"
		if codexOpen {
			writeOpenBindJSON(opts.Home, opts.AgentSessionID, openBindState{
				State:     "in_progress",
				StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
			})
		}

		type openBG struct {
			sid string
		}
		bgCh := make(chan openBG, 1)
		go func() {
			_ = waitForOpenReady(ctx, listenAddr, sessionID, 3*time.Second)
			if codexOpen {
				acceptCodexTrustRemote(ctx, listenAddr, sessionID, provider.BannerProvider, 15*time.Second, agentGone)
			}
			select {
			case <-ctx.Done():
				bgCh <- openBG{}
				return
			case <-time.After(200 * time.Millisecond):
			}
			if promptText != "" && (isResume || opts.NoSubmit || runnerID == "codex-tty") {
				if codexOpen {
					waitForOpenComposer(ctx, listenAddr, sessionID, provider.BannerProvider, openComposerWaitTimeout, agentGone)
				} else {
					_ = waitForBannerRemoteOpts(ctx, listenAddr, sessionID, provider.BannerProvider, provider.BannerMarkers, 0, agentGone)
				}
				if err := InjectMessage(listenAddr, sessionID, runnerID, promptText, !opts.NoSubmit); err != nil {
					recordSessionError(opts, "prompt-inject", err)
				}
			}
			sid := resolveOpenDetachRunnerSessionID(ctx, opts, runnerID, provider, configHome, runStart, listenAddr, sessionID)
			if sid != "" {
				writeOpenBindJSON(opts.Home, opts.AgentSessionID, openBindState{
					State:           "ok",
					StartedAt:       time.Now().UTC().Format(time.RFC3339Nano),
					FinishedAt:      time.Now().UTC().Format(time.RFC3339Nano),
					RunnerSessionID: sid,
				})
				if opts.OnRunnerSessionID != nil {
					opts.OnRunnerSessionID(sid)
				}
			} else if codexOpen {
				writeOpenBindJSON(opts.Home, opts.AgentSessionID, openBindState{
					State:      "failed",
					FinishedAt: time.Now().UTC().Format(time.RFC3339Nano),
					Error:      "codex session id not resolved",
				})
			}
			bgCh <- openBG{sid: sid}
		}()

		_, attachErr := ttywatch.AttachWriter(listenAddr, sessionID, attachMode)
		var runnerSID string
		if attachErr != nil {
			select {
			case bg := <-bgCh:
				runnerSID = bg.sid
			case <-time.After(100 * time.Millisecond):
			}
			return runnerSID, terminalSessionID, attachErr
		}
		bg := <-bgCh
		runnerSID = bg.sid
		if strings.TrimSpace(runnerSID) == "" && codexOpen {
			if id := tryDiscoverCodexOnce(opts, configHome, runStart, listenAddr, sessionID); id != "" {
				runnerSID = id
				if opts.OnRunnerSessionID != nil {
					opts.OnRunnerSessionID(id)
				}
			}
		}
		return runnerSID, terminalSessionID, nil
	}

	// Non-open headless: hard-wait inject-ready (banner/OpenReady), then inject.
	if err := waitForBannerRemote(ctx, listenAddr, sessionID, provider.BannerProvider, provider.BannerMarkers); err != nil {
		return "", terminalSessionID, err
	}

	select {
	case <-ctx.Done():
		return "", terminalSessionID, ctx.Err()
	case <-time.After(200 * time.Millisecond):
	}

	// Grok new session: prompt already on argv (auto-submits) — do not re-inject
	// the prompt text. Re-injecting under AGENT_RUN_GROK_TTY_COMMAND broke
	// no-double-inject / delayed-banner doctests and raced short-lived fakes
	// ("inject endpoint not found" once the PTY exited). codex-tty: always
	// inject (argv omitted above); resume/NoSubmit inject for all.
	shouldInject := promptText != "" && (isResume || opts.NoSubmit || runnerID == "codex-tty")
	usesGrokTTYHook := runnerID == "grok-tty" && strings.TrimSpace(os.Getenv(envGrokTTYCommand)) != ""
	usesLLMMockGrokHook := runnerID == "grok-tty" && strings.TrimSpace(os.Getenv("LLM_MOCK_RUN_GROK_COMMAND")) != ""
	if shouldInject {
		if err := InjectMessage(listenAddr, sessionID, runnerID, promptText, !opts.NoSubmit); err != nil {
			// Soft-skip when KeepAlive kept the HTTP serve up after the PTY
			// child already exited (resume-then-sleep fakes). If the agent is
			// still alive, hard-fail — soft-skip would hang discovery.
			agentExited := RegistryAgentExited(opts.Home, runnerID, sessionID)
			serveGone := !ttywatch.TCPReachable(listenAddr)
			if !(opts.KeepTerminalAlive && injectSessionGone(err) && (serveGone || agentExited)) {
				return "", terminalSessionID, err
			}
			fmt.Fprintf(opts.Stderr, "%s: inject skipped (pty gone under keep-alive): %v\n", runnerID, err)
		}
	} else if !opts.KeepTerminalAlive || usesGrokTTYHook || usesLLMMockGrokHook {
		// Soft kick: bare newline unblocks fake TUIs that `read` after the
		// banner without retyping the argv prompt (not a double-submit).
		// KeepAlive + real grok (no hook): skip — avoid spurious Enter.
		_ = ttywatch.InjectInput(listenAddr, sessionID, []byte("\n"))
	}

	keepBlocking := false
	if runnerID == "stub-tty" && opts.KeepTerminalAlive {
		if sc := loadStubScenario(); sc != nil && !sc.ExitAfterTurn {
			keepBlocking = true
		}
	}
	if keepBlocking {
		select {}
	}

	var autoExitCancel context.CancelFunc
	if isCodexProvider(provider.BannerProvider) || runnerID == "codex-tty" {
		// NoSubmit must not force Enter via retry submit path.
		// Retry Enter when prompt may still sit in the composer (inject or argv draft).
		if promptText != "" && !opts.NoSubmit {
			go retryCodexSubmitRemote(ctx, listenAddr, sessionID, promptText)
		}
		if !opts.KeepTerminalAlive {
			autoExitCtx, cancel := context.WithCancel(ctx)
			autoExitCancel = cancel
			go autoExitCodexAfterTurnRemote(autoExitCtx, listenAddr, sessionID, promptText)
		}
	}
	if runnerID == "commandcode-tty" && !opts.KeepTerminalAlive && !opts.Open {
		// cmd -p exits non-interactively; auto-exit not needed.
	}

	var tailState struct {
		sync.Mutex
		streamed bool
		id       string
	}
	var tailCancel context.CancelFunc
	var tailWG sync.WaitGroup
	var turnCompleted chan struct{}

	if opts.Emit != nil && provider.StartEventTail != nil && runnerID == "stub-tty" {
		_, cancel := context.WithCancel(ctx)
		tailCancel = cancel
		tailWG.Add(1)
		go func() {
			defer tailWG.Done()
			id, tailErr := provider.StartEventTail(TailContext{
				ScenarioPath: os.Getenv(envStubTTYScenario),
				Emit:         opts.Emit,
			})
			if tailErr == nil {
				tailState.Lock()
				tailState.id = id
				tailState.Unlock()
			}
		}()
	}

	if opts.Emit != nil && !provider.DisableTail && runnerID == "grok-tty" && !opts.GrokSyncOwnsEvents {
		grokHome := GrokHomeForRunner(configHome)
		tailCtx, cancel := context.WithCancel(ctx)
		tailCancel = cancel
		if opts.KeepTerminalAlive {
			turnCompleted = make(chan struct{})
		}
		tailWG.Add(1)
		go func() {
			defer tailWG.Done()
			_ = opts.Emit(types.AgentEvent{
				Type:      types.ActionThink,
				Text:      "Resolve session id...",
				Timestamp: time.Now().UnixMilli(),
			})
			id, updatesPath, discErr := DiscoverSession(tailCtx, grokHome, opts.Workspace, promptText, runStart)
			if discErr != nil {
				_ = opts.Emit(types.AgentEvent{
					Type:      types.ActionError,
					Text:      "Cannot resolve session id: " + discErr.Error(),
					Timestamp: time.Now().UnixMilli(),
				})
				recordSessionError(opts, "grok-session-discovery", discErr)
				return
			}
			if absUpdates, absErr := filepath.Abs(updatesPath); absErr == nil {
				updatesPath = absUpdates
			}
			fmt.Fprintf(opts.Stderr, "grok-tty: grok session %s\n", id)
			fmt.Fprintf(opts.Stderr, "grok-tty: grok updates %s\n", updatesPath)
			tailState.Lock()
			tailState.id = id
			tailState.Unlock()
			_ = opts.Emit(types.AgentEvent{
				Type:      types.ActionThink,
				Text:      "buffering grok session updates",
				Timestamp: time.Now().UnixMilli(),
			})
			startOffset := updatesTailStartOffset(updatesPath, runStart)
			onStreamed := func() {
				tailState.Lock()
				tailState.streamed = true
				tailState.Unlock()
			}
			emit := opts.Emit
			if turnCompleted != nil {
				var turnDone bool
				emit = func(ev types.AgentEvent) error {
					if err := opts.Emit(ev); err != nil {
						return err
					}
					if !turnDone && ev.Type == types.ActionDone {
						turnDone = true
						close(turnCompleted)
					}
					return nil
				}
			}
			_ = TailUpdatesFromOffset(tailCtx, updatesPath, startOffset, emit, onStreamed)
		}()
	}

	if opts.Emit != nil && runnerID == "codex-tty" {
		codexHome := CodexHomeForRunner(configHome)
		tailCtx, cancel := context.WithCancel(ctx)
		tailCancel = cancel
		tailWG.Add(1)
		go func() {
			defer tailWG.Done()
			var codexSessionID string
			var transcriptPath string
			poll := time.NewTicker(tailPollInterval)
			defer poll.Stop()
			for transcriptPath == "" {
				snapshot, snapErr := fetchSnapshotBytes(listenAddr, sessionID)
				if snapErr == nil {
					codexSessionID = FindCodexResumeSessionID(string(snapshot))
					if codexSessionID != "" {
						path, ok, discErr := FindCodexTranscriptBySessionID(codexHome, codexSessionID)
						if discErr != nil {
							recordSessionError(opts, "codex-transcript-discovery", discErr)
							return
						}
						if ok {
							transcriptPath = path
							break
						}
					}
				}
				id, path, ok, discErr := scanActiveCodexTranscripts(codexHome, opts.Workspace, runStart, promptText)
				if discErr != nil {
					recordSessionError(opts, "codex-transcript-scan", discErr)
					return
				}
				if ok {
					codexSessionID = id
					transcriptPath = path
					break
				}
				select {
				case <-tailCtx.Done():
					return
				case <-poll.C:
				}
			}
			if absTranscript, absErr := filepath.Abs(transcriptPath); absErr == nil {
				transcriptPath = absTranscript
			}
			tailState.Lock()
			tailState.id = codexSessionID
			tailState.streamed = true
			tailState.Unlock()
			_ = TailCodexTranscriptFromOffset(tailCtx, transcriptPath, 0, opts.Emit)
		}()
	}

	var waitErr error
	var snapshotHold []byte
	var lastSnapMu sync.Mutex
	// Poll snapshots while waiting so scrollback fallback still works when the
	// PTY child exits under KeepAlive (serve stays up; session /input may 404).
	holdSnap := func() {
		if snap, err := fetchSnapshotBytes(listenAddr, sessionID); err == nil && len(snap) > 0 {
			lastSnapMu.Lock()
			snapshotHold = snap
			lastSnapMu.Unlock()
		}
	}
	if opts.KeepTerminalAlive {
		if strings.TrimSpace(promptText) == "" {
			waitErr = nil
		} else {
			var extraComplete func() bool
			if runnerID == "grok-tty" {
				if !opts.GrokSyncOwnsEvents {
					extraComplete = func() bool {
						tailState.Lock()
						defer tailState.Unlock()
						return tailState.streamed
					}
				} else {
					// Sync owns discovery/events: settle when runner_session_id is
					// bound or a resolve error was written (never via PTY chrome).
					agentHome := opts.Home
					agentSID := opts.AgentSessionID
					extraComplete = func() bool {
						return grokKeepTTYTurnSettled(agentHome, agentSID)
					}
				}
			} else if runnerID == "codex-tty" || runnerID == "commandcode-tty" {
				// KeepAlive leaves the HTTP serve up after the PTY child exits.
				// Settle when: resume/exit footer is on screen (resume-sleep
				// fakes print it before sleep), agent PID is dead, or registry
				// command_exited is set. Footer is the fast path — waiting only
				// on registry/PID burned stream-probe ExecTimeout under CI load.
				home := opts.Home
				rid, sid := runnerID, sessionID
				addr := listenAddr
				commandPID := 0
				if result.Entry != nil {
					commandPID = result.Entry.CommandPID
				}
				extraComplete = func() bool {
					if snap, err := fetchSnapshotBytes(addr, sid); err == nil &&
						ScrollbackSuggestsAgentExited(string(snap), rid) {
						return true
					}
					if commandPID > 0 && !ttywatch.ProcessAlive(commandPID) {
						return true
					}
					return RegistryAgentExited(home, rid, sid)
				}
			}
			waitDone := make(chan error, 1)
			go func() {
				waitDone <- waitForPersistentTurnRemote(ctx, listenAddr, sessionID, promptText, cfg, extraComplete)
			}()
			// 250ms: avoid stampeding ptywrap snapshots with the codex tail
			// goroutine under parallel CI load (stream-probe flakes).
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()
		keepWaitLoop:
			for {
				select {
				case waitErr = <-waitDone:
					holdSnap()
					break keepWaitLoop
				case <-ticker.C:
					holdSnap()
				}
			}
		}
	} else {
		// Without KeepAlive the serve exits with the PTY child, so a snapshot
		// after WaitHeadless is connection-refused. Poll while waiting and keep
		// the last good frame for scrollback fallback (fake-TUI doctests).
		waitDone := make(chan error, 1)
		go func() {
			waitDone <- ttywatch.WaitHeadless(ctx, result, argv)
		}()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
	waitLoop:
		for {
			select {
			case waitErr = <-waitDone:
				holdSnap()
				break waitLoop
			case <-ticker.C:
				holdSnap()
			}
		}
		if autoExitCancel != nil {
			autoExitCancel()
		}
	}

	if runnerID == "grok-tty" && opts.KeepTerminalAlive && turnCompleted != nil {
		select {
		case <-turnCompleted:
		case <-ctx.Done():
		}
	}

	if tailCancel != nil {
		tailCancel()
		tailWG.Wait()
	}

	snapshot, snapErr := fetchSnapshotBytes(listenAddr, sessionID)
	lastSnapMu.Lock()
	held := append([]byte(nil), snapshotHold...)
	lastSnapMu.Unlock()
	if (snapErr != nil || len(snapshot) == 0) && len(held) > 0 {
		snapshot, snapErr = held, nil
	}
	captured := ""
	if snapErr == nil {
		captured = extractAssistantTextForProvider(snapshot, opts.Prompt, provider.BannerMarkers, provider.BannerProvider)
	}

	tailState.Lock()
	runnerSessionID = tailState.id
	streamed := tailState.streamed
	tailState.Unlock()

	// Scrollback fallback when no provider transcript streamed. keep-tty grok must
	// not treat PTY chrome as assistant output; codex/commandcode headless always
	// set KeepAlive, so fake-TUI hooks still need fallback under keep-alive.
	allowScrollbackFallback := !opts.KeepTerminalAlive ||
		(runnerID == "codex-tty" && strings.TrimSpace(os.Getenv(envCodexTTYCommand)) != "") ||
		runnerID == "commandcode-tty"
	if !streamed && opts.Emit != nil && allowScrollbackFallback && (runnerID == "codex-tty" || runnerID == "commandcode-tty" || runnerID == "grok-tty") {
		if runnerID == "codex-tty" {
			fmt.Fprintf(opts.Stderr, "codex-tty: codex transcript not found; falling back to scrollback capture\n")
		} else if runnerID == "grok-tty" {
			fmt.Fprintf(opts.Stderr, "grok-tty: grok session not found; falling back to scrollback capture\n")
		}
		text := strings.TrimSpace(captured)
		if text != "" {
			if emitErr := opts.Emit(types.AgentEvent{
				Type:      types.ActionMessage,
				Role:      "assistant",
				Text:      text,
				Timestamp: time.Now().UnixMilli(),
			}); emitErr != nil {
				return runnerSessionID, terminalSessionID, emitErr
			}
		}
	}

	if terminalSessionID != "" && opts.AgentSessionID != "" && result.Entry != nil {
		screenStatus := "unknown"
		if provider.DetectScreenStatus != nil && snapErr == nil {
			screenStatus = provider.DetectScreenStatus(snapshot)
		}
		_ = WriteTTYJSON(opts.Home, TTYSnapshot{
			RunnerID:          runnerID,
			AgentSessionID:    opts.AgentSessionID,
			TerminalSessionID: terminalSessionID,
			ListenAddr:        result.Entry.ListenAddr,
			PID:               result.Entry.PID,
			CreatedAt:         result.Entry.CreatedAt,
			ScreenStatus:      screenStatus,
			Alive:             opts.KeepTerminalAlive || waitErr == nil,
		})
	}

	if waitErr != nil {
		return runnerSessionID, terminalSessionID, waitErr
	}
	return runnerSessionID, terminalSessionID, nil
}

// waitForOpenReady polls until scrollback has any content or timeout elapses.
// Used on the --open attach-first path (including AGENT_RUN_OPEN_ATTACH_INSTANT)
// as a best-effort soft wait; callers ignore timeout so open still attaches.
func waitForOpenReady(ctx context.Context, listenAddr, sessionID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		snapshot, err := ttywatch.SnapshotText(listenAddr, sessionID)
		if err == nil && strings.TrimSpace(snapshot) != "" {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("open ready: scrollback empty after %s", timeout)
}

// resolveOpenDetachRunnerSessionID returns the provider session id to bind on
// --open / --detach return. Resume reuses ResumeSessionID. Codex new sessions
// wait briefly for a cwd-matched rollout (or resume footer) so agent-run meta
// can ModeResume after terminal close. Non-codex open returns empty (grok uses
// a separate open-bind worker in agentui).
func resolveOpenDetachRunnerSessionID(ctx context.Context, opts RunOptions, runnerID string, provider Provider, configHome string, runStart time.Time, listenAddr, termSessionID string) string {
	if id := strings.TrimSpace(opts.ResumeSessionID); id != "" {
		return id
	}
	if !isCodexProvider(provider.BannerProvider) && runnerID != "codex-tty" {
		return ""
	}
	codexHome := CodexHomeForRunner(configHome)
	prompt := prepareRunnerPrompt(runnerID, strings.TrimSpace(opts.Prompt))
	sid := WaitDiscoverCodexSessionID(ctx, codexHome, opts.Workspace, runStart, listenAddr, termSessionID, prompt, openCodexBindBudget)
	if sid == "" {
		return ""
	}
	return sid
}

func tryDiscoverCodexOnce(opts RunOptions, configHome string, runStart time.Time, listenAddr, termSessionID string) string {
	scrollback := ""
	if strings.TrimSpace(listenAddr) != "" && strings.TrimSpace(termSessionID) != "" {
		if snap, err := fetchSnapshotBytes(listenAddr, termSessionID); err == nil {
			scrollback = string(snap)
		}
	}
	prompt := prepareRunnerPrompt("codex-tty", strings.TrimSpace(opts.Prompt))
	id, ok := DiscoverCodexSessionID(CodexHomeForRunner(configHome), opts.Workspace, runStart, scrollback, prompt)
	if !ok {
		return ""
	}
	return id
}

// writeIdlePolicyBeforeServe persists launch-time idle-exit for the serve child.
func writeIdlePolicyBeforeServe(opts RunOptions) error {
	if !opts.ExitOnIdle {
		return nil
	}
	if opts.IdleTimeout < 0 {
		return nil
	}
	home := strings.TrimSpace(os.Getenv("AGENT_RUN_HOME"))
	if home == "" {
		home = strings.TrimSpace(opts.Home)
	}
	sessionID := strings.TrimSpace(opts.AgentSessionID)
	if sessionID == "" {
		sessionID = strings.TrimSpace(opts.SessionID)
	}
	if home == "" || sessionID == "" {
		return nil
	}
	return agentstorage.WriteIdlePolicy(home, sessionID, agentstorage.IdlePolicy{
		ExitOnIdle:  true,
		IdleTimeout: opts.IdleTimeout,
	})
}

// grokKeepTTYTurnSettled reports whether keep-tty + sync-owned discovery has
// either bound a runner_session_id or recorded a resolve error in events.jsonl.
func grokKeepTTYTurnSettled(home, agentSessionID string) bool {
	home = strings.TrimSpace(home)
	agentSessionID = strings.TrimSpace(agentSessionID)
	if home == "" || agentSessionID == "" {
		return false
	}
	sessionDir := filepath.Join(home, "sessions", agentSessionID)
	metaPath := filepath.Join(sessionDir, "meta.json")
	if data, err := os.ReadFile(metaPath); err == nil {
		var meta struct {
			RunnerSessionID string `json:"runner_session_id"`
		}
		if json.Unmarshal(data, &meta) == nil && strings.TrimSpace(meta.RunnerSessionID) != "" {
			return true
		}
	}
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev struct {
			Type string `json:"type"`
			Text string `json:"text"`
			Role string `json:"role"`
		}
		if json.Unmarshal([]byte(line), &ev) != nil {
			continue
		}
		if ev.Type == "error" && strings.Contains(ev.Text, "Cannot resolve session id:") {
			return true
		}
		if ev.Type == "message" && ev.Role == "assistant" && strings.TrimSpace(ev.Text) != "" {
			return true
		}
		if ev.Type == "done" {
			return true
		}
	}
	return false
}
