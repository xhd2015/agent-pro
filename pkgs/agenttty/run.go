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
	Home                  string
	Workspace             string
	Prompt                string
	Model                 string
	ResumeSessionID       string
	RunnerID              string
	AgentSessionID        string
	SettingsPath          string
	AgentPath             string
	AgentRunnerConfigHome string
	BinaryPath            string
	KeepTerminalAlive     bool
	Stderr                io.Writer
	Emit                  func(types.AgentEvent) error
	OnTerminalSessionID   func(string)
}

// RunHeadless starts a detached ttywatch serve session, injects the prompt, tails
// agent events, and waits for the child to exit unless KeepTerminalAlive is set.
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
	if strings.TrimSpace(opts.Prompt) == "" {
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

	configHome := ResolveAgentRunnerConfigHome(opts.AgentRunnerConfigHome)
	if configHome == "" {
		if provisioned, provErr := AutoProvisionGrokConfigHome(runnerID, opts.AgentPath, ""); provErr != nil {
			return "", "", provErr
		} else if provisioned != "" {
			configHome = provisioned
		}
	}
	argv = PrependCommandEnv(argv, runnerID, configHome)

	binaryPath, err := resolveBinaryPath(opts.BinaryPath)
	if err != nil {
		return "", "", err
	}

	result, err := ttywatch.HeadlessRun(ctx, ttywatch.HeadlessRunOptions{
		Home:           opts.Home,
		RegistrySubdir: provider.RegistryDir,
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

	fmt.Fprintf(opts.Stderr, "%s: %s\n", runnerID, sessionID)

	runStart := time.Now()
	promptText := strings.TrimSpace(opts.Prompt)

	if err := waitForBannerRemote(ctx, listenAddr, sessionID, provider.BannerProvider, provider.BannerMarkers); err != nil {
		return "", terminalSessionID, err
	}

	select {
	case <-ctx.Done():
		return "", terminalSessionID, ctx.Err()
	case <-time.After(200 * time.Millisecond):
	}

	if err := ttywatch.SendMessage(listenAddr, sessionID, promptText, true); err != nil {
		return "", terminalSessionID, err
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
		go retryCodexSubmitRemote(ctx, listenAddr, sessionID, promptText)
		if !opts.KeepTerminalAlive {
			autoExitCtx, cancel := context.WithCancel(ctx)
			autoExitCancel = cancel
			go autoExitCodexAfterTurnRemote(autoExitCtx, listenAddr, sessionID, promptText)
		}
	}

	var tailState struct {
		sync.Mutex
		streamed bool
		id       string
	}
	var tailCancel context.CancelFunc
	var tailWG sync.WaitGroup

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

	if opts.Emit != nil && !provider.DisableTail && runnerID == "grok-tty" {
		grokHome := GrokHomeForRunner(configHome)
		tailCtx, cancel := context.WithCancel(ctx)
		tailCancel = cancel
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
			_ = TailUpdatesFromOffset(tailCtx, updatesPath, startOffset, opts.Emit, onStreamed)
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
		var extraComplete func() bool
		if runnerID == "grok-tty" {
			extraComplete = func() bool {
				tailState.Lock()
				defer tailState.Unlock()
				return tailState.streamed
			}
		}
		waitErr = waitForPersistentTurnRemote(ctx, listenAddr, sessionID, promptText, cfg, extraComplete)
	} else {
		waitErr = ttywatch.WaitHeadless(ctx, result, argv)
		if autoExitCancel != nil {
			autoExitCancel()
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

	if !streamed && opts.Emit != nil && runnerID == "codex-tty" {
		fmt.Fprintf(opts.Stderr, "codex-tty: codex transcript not found; falling back to scrollback capture\n")
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