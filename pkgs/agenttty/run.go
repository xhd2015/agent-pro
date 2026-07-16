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
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

// RunOptions configures a headless TTY runner invocation via detached serve.
type RunOptions struct {
	Home            string
	Workspace       string
	Prompt          string
	Model           string
	ResumeSessionID string
	RunnerID        string
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
	Env               []string
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
	if opts.Open || opts.Detach {
		opts.KeepTerminalAlive = true
	}
	// commandcode-tty: always keep PTY alive so snapshot/attach work after run.
	// Headless mode injects -p (non-interactive) so cmd exits cleanly while the
	// serve persists the scrollback. Open mode omits -p for interactive use.
	if runnerID == "commandcode-tty" {
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
	// New session (no --resume): pass initial prompt as trailing positional arg
	// (grok [PROMPT]). Resume follow-ups stay inject-only so argv keeps --resume.
	// NoSubmit: never put draft on argv (real Grok auto-submits positional PROMPT).
	// Headless still injects after banner for turn completion with fake TUI scripts.
	// commandcode-tty headless: inject -p so cmd runs in non-interactive print mode
	// and works with the mock server; open mode omits -p for interactive use.
	if strings.TrimSpace(opts.ResumeSessionID) == "" && !opts.NoSubmit {
		if p := strings.TrimSpace(opts.Prompt); p != "" {
			if runnerID == "commandcode-tty" && !opts.Open {
				argv = append(argv, "-p", p)
			} else {
				argv = append(argv, p)
			}
		}
	}

	configHome := ResolveAgentRunnerConfigHome(opts.AgentRunnerConfigHome)
	if configHome == "" {
		if provisioned, provErr := AutoProvisionGrokConfigHome(runnerID, opts.AgentPath, ""); provErr != nil {
			return "", "", provErr
		} else if provisioned != "" {
			configHome = provisioned
		}
	}
	argv = ApplyChildProcessEnv(argv, runnerID, configHome, opts.PrependPaths, opts.Env)

	binaryPath, err := resolveBinaryPath(opts.BinaryPath)
	if err != nil {
		return "", "", err
	}

	result, err := ttywatch.HeadlessRun(ctx, ttywatch.HeadlessRunOptions{
		Home:           opts.Home,
		RegistrySubdir: provider.RegistryDir,
		SessionID:      strings.TrimSpace(opts.SessionID),
		BinaryPath:     binaryPath,
		Command:        argv,
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
	promptText := strings.TrimSpace(opts.Prompt)
	isResume := strings.TrimSpace(opts.ResumeSessionID) != ""

	// --detach: keep-alive daemon only — no attach, no event stream, no wait for turn.
	// Prompt on argv for new sessions when non-empty and not NoSubmit (same as non-open).
	// Resume follow-up injects after soft banner wait; empty prompt is reopen-only.
	if opts.Detach {
		if promptText != "" && (isResume || opts.NoSubmit) {
			if readyErr := waitForBannerRemote(ctx, listenAddr, sessionID, provider.BannerProvider, provider.BannerMarkers); readyErr != nil {
				if ctx.Err() != nil {
					return "", terminalSessionID, ctx.Err()
				}
				// Soft: skip inject on banner timeout; still leave daemon alive.
			} else if err := ttywatch.SendMessage(listenAddr, sessionID, promptText, !opts.NoSubmit); err != nil {
				return "", terminalSessionID, err
			}
		} else {
			// Brief soft ready so registry/PTY is fully up before parent exits.
			_ = waitForOpenReady(ctx, listenAddr, sessionID, 3*time.Second)
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
		// Leave PTY/registry alive for later attach/send; return immediately.
		return "", terminalSessionID, nil
	}

	// --open is attach-first: banner/OpenReady is inject-readiness, not attach readiness.
	// Soft-wait for any scrollback content; never fail open on ready-marker timeout.
	// AGENT_RUN_OPEN_ATTACH_INSTANT=1 keeps the same short soft wait (AttachWriter also
	// returns immediately under that env).
	if opts.Open {
		_ = waitForOpenReady(ctx, listenAddr, sessionID, 3*time.Second)

		select {
		case <-ctx.Done():
			return "", terminalSessionID, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}

		// Inject policy under --open:
		// - New session (default): initial prompt already on argv when non-empty → do not re-inject.
		// - New session + NoSubmit: draft not on argv → inject with suffixCR=false.
		// - Resume + non-empty follow-up: wait inject-ready then inject if ready; banner
		//   timeout must not fail open — still attach.
		// - Empty prompt: no inject.
		// - NoSubmit: when inject happens, suffixCR=false.
		if promptText != "" && (isResume || opts.NoSubmit) {
			if readyErr := waitForBannerRemote(ctx, listenAddr, sessionID, provider.BannerProvider, provider.BannerMarkers); readyErr != nil {
				if ctx.Err() != nil {
					return "", terminalSessionID, ctx.Err()
				}
				// Soft: skip inject on banner/OpenReady timeout; proceed to attach.
			} else if err := ttywatch.SendMessage(listenAddr, sessionID, promptText, !opts.NoSubmit); err != nil {
				return "", terminalSessionID, err
			}
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
		if _, attachErr := ttywatch.AttachWriter(listenAddr, sessionID, "attach"); attachErr != nil {
			return "", terminalSessionID, attachErr
		}
		// Leave PTY/registry alive for re-attach/send; do not wait for turn.
		return "", terminalSessionID, nil
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

	// New session already has prompt on argv (real Grok auto-submits).
	// Re-injecting double-submits. Inject only for resume follow-up or NoSubmit.
	// NoSubmit: inject without trailing Enter (suffixCR=false).
	shouldInject := promptText != "" && (isResume || opts.NoSubmit)
	if shouldInject {
		if err := ttywatch.SendMessage(listenAddr, sessionID, promptText, !opts.NoSubmit); err != nil {
			return "", terminalSessionID, err
		}
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
		// Argv-only new sessions: do not force re-submit (would double-submit).
		if shouldInject && !opts.NoSubmit {
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

func resolveBinaryPath(explicit string) (string, error) {
	if v := strings.TrimSpace(explicit); v != "" {
		return v, nil
	}
	if len(os.Args) > 0 && strings.TrimSpace(os.Args[0]) != "" {
		return os.Args[0], nil
	}
	return os.Executable()
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
