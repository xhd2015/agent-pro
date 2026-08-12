package agenttty

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// RunOptions configures a headless TTY runner invocation via detached serve.
type RunOptions struct {
	Home            string
	Workspace       string
	Prompt          string
	Model           string
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
	// commandcode-tty: keep PTY alive for headless snapshot/attach after run.
	// Open + OpenCloseExits: do not re-force keep-alive (window-close teardown).
	if runnerID == "commandcode-tty" {
		if !(opts.Open && OpenCloseExits()) {
			opts.KeepTerminalAlive = true
		}
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
	promptText := normalizeRunnerPrompt(strings.TrimSpace(opts.Prompt))
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

	// --open: soft-wait inject-readiness, inject when ready, then AttachWriter.
	// AGENT_RUN_OPEN_ATTACH_INSTANT=1 skips interactive attach (tests).
	// Abort pre-attach waits when the PTY agent child already exited so resume
	// reopen does not hang blank for minutes then AttachWriter forever.
	if opts.Open {
		agentGone := func() bool {
			return RegistryAgentExited(opts.Home, runnerID, sessionID)
		}
		_ = waitForOpenReady(ctx, listenAddr, sessionID, 3*time.Second)
		// Codex open: soft-accept directory trust so follow-up send is not blocked.
		// Short no-trust grace (see acceptCodexTrustRemote) avoids a 45s blank hang.
		if isCodexProvider(provider.BannerProvider) || runnerID == "codex-tty" {
			acceptCodexTrustRemote(ctx, listenAddr, sessionID, provider.BannerProvider, 15*time.Second, agentGone)
		}

		select {
		case <-ctx.Done():
			return "", terminalSessionID, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}

		if agentGone() {
			return "", terminalSessionID, formatOpenAgentExitedError(runnerID, listenAddr, sessionID, "open attach", nil)
		}

		// Inject policy under --open:
		// - Grok new session: prompt on argv → do not re-inject.
		// - codex-tty: always inject after ready (no argv prompt).
		// - New session + NoSubmit: draft inject without Enter.
		// - Resume + non-empty follow-up: wait inject-ready then inject if ready; banner
		//   timeout must not fail open — still attach when agent is alive.
		// - Empty prompt: no inject.
		if promptText != "" && (isResume || opts.NoSubmit || runnerID == "codex-tty") {
			bannerTimeout := time.Duration(0) // default
			if isCodexProvider(provider.BannerProvider) || runnerID == "codex-tty" {
				bannerTimeout = openCodexBannerWaitTimeout
			}
			readyErr := waitForBannerRemoteOpts(ctx, listenAddr, sessionID, provider.BannerProvider, provider.BannerMarkers, bannerTimeout, agentGone)
			if readyErr != nil {
				if ctx.Err() != nil {
					return "", terminalSessionID, ctx.Err()
				}
				if agentGone() {
					return "", terminalSessionID, formatOpenAgentExitedError(runnerID, listenAddr, sessionID, "inject-ready", readyErr)
				}
				// Soft: skip inject on banner timeout; proceed to attach if still alive.
			} else if err := InjectMessage(listenAddr, sessionID, runnerID, promptText, !opts.NoSubmit); err != nil {
				return "", terminalSessionID, err
			}
		}

		if agentGone() {
			return "", terminalSessionID, formatOpenAgentExitedError(runnerID, listenAddr, sessionID, "open attach", nil)
		}

		// Persist dual-write snapshot while keep-alive, then auto-attach.
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
		// OpenCloseExits (default): attach_mode=open → ptywrap roleWriter so bare
		// WS disconnect (iTerm red-close) calls stopChild() without detach_keep,
		// while the initial frame is raw scrollback (attach-like) so Grok mouse /
		// alt-screen CSIs reach the host. attach_mode=attach (roleAttacher) leaves
		// ghost __serve__; attach_mode=screen is CUP export for grid tools only.
		attachMode := "attach"
		if OpenCloseExits() {
			attachMode = "open"
		}
		// Bind codex runner_session_id before attach returns. Open path used to
		// always return "" here; without a bound id, AutoSendOrResume falls to
		// ModeRun after terminal close and opens a second Codex conversation.
		runnerSID := resolveOpenDetachRunnerSessionID(ctx, opts, runnerID, provider, configHome, runStart, listenAddr, sessionID)

		if _, attachErr := ttywatch.AttachWriter(listenAddr, sessionID, attachMode); attachErr != nil {
			return runnerSID, terminalSessionID, attachErr
		}
		// With OpenCloseExits + !KeepTerminalAlive, window close reaps child and
		// serve exits. Ctrl-] still sends detach_keep (child kept).
		// One-shot re-try if still empty (rollout may appear during attach).
		if strings.TrimSpace(runnerSID) == "" && (isCodexProvider(provider.BannerProvider) || runnerID == "codex-tty") {
			if id := tryDiscoverCodexOnce(opts, configHome, runStart, listenAddr, sessionID); id != "" {
				runnerSID = id
				fmt.Fprintf(opts.Stderr, "codex-tty: codex session %s\n", runnerSID)
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
	// the prompt text. codex-tty: always inject (argv omitted above); resume/NoSubmit
	// inject for all. NoSubmit: inject without trailing Enter (suffixCR=false).
	// codex-tty uses InjectMessage (type then separate Enter) when submitting.
	shouldInject := promptText != "" && (isResume || opts.NoSubmit || runnerID == "codex-tty")
	if shouldInject {
		if err := InjectMessage(listenAddr, sessionID, runnerID, promptText, !opts.NoSubmit); err != nil {
			return "", terminalSessionID, err
		}
	} else if !opts.KeepTerminalAlive {
		// Soft kick for shell env-logger fakes that `read` after printing the
		// banner (session-env / run-color). Real grok already received the
		// prompt on argv; a bare newline only unblocks read without retyping.
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
				fmt.Fprintf(opts.Stderr, "grok-tty: grok session discovery failed: %v\n", discErr)
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
						path, ok, discErr := findCodexTranscriptBySessionID(codexHome, codexSessionID)
						if discErr != nil {
							if tailCtx.Err() == nil {
								fmt.Fprintf(opts.Stderr, "codex-tty: codex transcript discovery failed: %v\n", discErr)
							}
							return
						}
						if ok {
							transcriptPath = path
							break
						}
					}
				}
				id, path, ok, discErr := scanActiveCodexTranscripts(codexHome, opts.Workspace, runStart)
				if discErr != nil {
					if tailCtx.Err() == nil {
						fmt.Fprintf(opts.Stderr, "codex-tty: codex transcript discovery failed: %v\n", discErr)
					}
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
			fmt.Fprintf(opts.Stderr, "codex-tty: codex session %s\n", codexSessionID)
			fmt.Fprintf(opts.Stderr, "codex-tty: codex transcript %s\n", transcriptPath)
			tailState.Lock()
			tailState.id = codexSessionID
			tailState.streamed = true
			tailState.Unlock()
			_ = TailCodexTranscriptFromOffset(tailCtx, transcriptPath, 0, opts.Emit)
		}()
	}

	var waitErr error
	if opts.KeepTerminalAlive {
		if strings.TrimSpace(promptText) == "" {
			waitErr = nil
		} else {
			var extraComplete func() bool
			if runnerID == "grok-tty" && !opts.GrokSyncOwnsEvents {
				extraComplete = func() bool {
					tailState.Lock()
					defer tailState.Unlock()
					return tailState.streamed
				}
			}
			waitErr = waitForPersistentTurnRemote(ctx, listenAddr, sessionID, promptText, cfg, extraComplete)
		}
	} else {
		waitErr = ttywatch.WaitHeadless(ctx, result, argv)
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
	captured := ""
	if snapErr == nil {
		captured = extractAssistantTextForProvider(snapshot, opts.Prompt, provider.BannerMarkers, provider.BannerProvider)
	}

	tailState.Lock()
	runnerSessionID = tailState.id
	streamed := tailState.streamed
	tailState.Unlock()

	if !streamed && opts.Emit != nil && (runnerID == "codex-tty" || runnerID == "commandcode-tty") {
		if runnerID == "codex-tty" {
			fmt.Fprintf(opts.Stderr, "codex-tty: codex transcript not found; falling back to scrollback capture\n")
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
	sid := WaitDiscoverCodexSessionID(ctx, codexHome, opts.Workspace, runStart, listenAddr, termSessionID, openCodexBindBudget)
	if sid == "" {
		return ""
	}
	fmt.Fprintf(opts.Stderr, "codex-tty: codex session %s\n", sid)
	return sid
}

func tryDiscoverCodexOnce(opts RunOptions, configHome string, runStart time.Time, listenAddr, termSessionID string) string {
	scrollback := ""
	if strings.TrimSpace(listenAddr) != "" && strings.TrimSpace(termSessionID) != "" {
		if snap, err := fetchSnapshotBytes(listenAddr, termSessionID); err == nil {
			scrollback = string(snap)
		}
	}
	id, ok := DiscoverCodexSessionID(CodexHomeForRunner(configHome), opts.Workspace, runStart, scrollback)
	if !ok {
		return ""
	}
	return id
}
