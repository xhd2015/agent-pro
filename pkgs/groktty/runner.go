package groktty

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

// RunOptions configures a grok-tty headless invocation.
type RunOptions struct {
	Home                string
	Workspace           string
	Prompt              string
	Model               string
	ResumeSessionID     string
	SettingsPath        string
	AgentPath           string
	RunnerID            string
	StderrPrefix        string
	RegistryDir         string
	BannerProvider      string
	BannerMarkers       []string
	DisableTail         bool
	KeepTerminalAlive   bool
	Stderr              io.Writer
	Emit                func(types.AgentEvent) error
	OnTerminalSessionID func(string)
}

// Run starts an adhoc ptywrap server, runs interactive grok in a PTY, injects the
// prompt after the TUI banner, tails the matching grok session updates.jsonl when
// found, and blocks until the session exits.
func Run(ctx context.Context, opts RunOptions) (captured string, grokSessionID string, err error) {
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

	env := newExecEnv()
	cfg := opts.withDefaults()
	argv, err := cfg.buildArgv(env, opts.SettingsPath, opts.AgentPath, opts.Model, opts.ResumeSessionID)
	if err != nil {
		return "", "", err
	}

	mgr := ptywrap.NewManager()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", err
	}
	listenAddr := ln.Addr().String()

	mux := http.NewServeMux()
	ptywrap.RegisterAPIWithManager(mux, mgr)
	srv := &http.Server{Handler: mux}
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- srv.Serve(ln)
	}()

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		<-serverDone
	}

	sessionID := ""
	var releaseSessionID func()
	if cfg.registryDir != "" {
		var reserveErr error
		sessionID, releaseSessionID, reserveErr = ReserveRegistrySessionID(opts.Home, cfg.registryDir)
		if reserveErr != nil {
			shutdown()
			return "", "", reserveErr
		}
	}
	sess, err := mgr.CreateCommandWithID(sessionID, cfg.runnerID, opts.Workspace, argv)
	if err != nil {
		if releaseSessionID != nil {
			releaseSessionID()
		}
		shutdown()
		return "", "", err
	}
	sessionID = sess.ID
	runStart := time.Now()

	if err := WriteRegistryFor(opts.Home, cfg.registryDir, RegistryEntry{
		SessionID:  sessionID,
		ListenAddr: listenAddr,
		PID:        os.Getpid(),
	}); err != nil {
		if releaseSessionID != nil {
			releaseSessionID()
		}
		mgr.Remove(sessionID)
		shutdown()
		return "", "", err
	}
	if releaseSessionID != nil {
		releaseSessionID()
	}
	if opts.OnTerminalSessionID != nil {
		opts.OnTerminalSessionID(sessionID)
	}
	if !opts.KeepTerminalAlive {
		defer RemoveRegistryFor(opts.Home, cfg.registryDir, sessionID)
	}

	fmt.Fprintf(opts.Stderr, "%s: %s\n", cfg.stderrPrefix, sessionID)

	if err := waitForBannerConfig(ctx, mgr, sessionID, cfg.bannerProvider, cfg.bannerMarkers); err != nil {
		mgr.Remove(sessionID)
		shutdown()
		return "", "", err
	}

	// Let the TUI finish rendering the input prompt after the banner.
	select {
	case <-ctx.Done():
		return "", "", ctx.Err()
	case <-time.After(200 * time.Millisecond):
	}

	// Interactive TUIs submit on carriage return (Enter), not bare LF.
	promptText := strings.TrimSpace(opts.Prompt)
	if writeErr := mgr.WriteInput(sessionID, []byte(promptText)); writeErr != nil {
		mgr.Remove(sessionID)
		shutdown()
		return "", "", writeErr
	}
	if writeErr := mgr.WriteInput(sessionID, []byte("\r")); writeErr != nil {
		mgr.Remove(sessionID)
		shutdown()
		return "", "", writeErr
	}
	var autoExitCancel context.CancelFunc
	if cfg.bannerProvider == "codex" || cfg.bannerProvider == "codex-tty" || cfg.runnerID == "codex-tty" {
		go retryCodexSubmit(ctx, mgr, sessionID, promptText)
		if !opts.KeepTerminalAlive {
			autoExitCtx, cancel := context.WithCancel(ctx)
			autoExitCancel = cancel
			go autoExitCodexAfterTurn(autoExitCtx, mgr, sessionID, promptText)
		}
	}

	var tailState struct {
		sync.Mutex
		streamed bool
		id       string
	}
	var tailCancel context.CancelFunc
	var tailWG sync.WaitGroup

	if opts.Emit != nil && !cfg.disableTail {
		grokHome := GrokHome()
		tailCtx, cancel := context.WithCancel(ctx)
		tailCancel = cancel
		tailWG.Add(1)
		go func() {
			defer tailWG.Done()
			id, updatesPath, discErr := DiscoverSession(tailCtx, grokHome, opts.Workspace, promptText, runStart)
			if discErr != nil {
				if tailCtx.Err() == nil {
					fmt.Fprintf(opts.Stderr, "grok-tty: grok session discovery failed: %v\n", discErr)
				}
				return
			}
			absUpdates, absErr := filepath.Abs(updatesPath)
			if absErr == nil {
				updatesPath = absUpdates
			}
			fmt.Fprintf(opts.Stderr, "grok-tty: grok session %s\n", id)
			fmt.Fprintf(opts.Stderr, "grok-tty: grok updates %s\n", updatesPath)
			tailState.Lock()
			tailState.id = id
			tailState.streamed = true
			tailState.Unlock()
			// Preamble stdout event after stderr diagnostics so streamed markers
			// appear after the grok session stderr block in ordering checks.
			_ = opts.Emit(types.AgentEvent{
				Type:      types.ActionThink,
				Text:      "buffering grok session updates",
				Timestamp: time.Now().UnixMilli(),
			})
			startOffset := updatesTailStartOffset(updatesPath, runStart)
			_ = TailUpdatesFromOffset(tailCtx, updatesPath, startOffset, opts.Emit)
		}()
	}
	if opts.Emit != nil && cfg.runnerID == "codex-tty" {
		codexHome := CodexHome()
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
				codexSessionID = FindCodexResumeSessionID(string(mgr.Scrollback(sessionID)))
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
			absTranscript, absErr := filepath.Abs(transcriptPath)
			if absErr == nil {
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
		waitErr = waitForPersistentTurn(ctx, mgr, sessionID, promptText, cfg)
	} else {
		waitErr = mgr.Wait(sessionID)
		if autoExitCancel != nil {
			autoExitCancel()
		}
	}
	scrollback := mgr.Scrollback(sessionID)
	if !opts.KeepTerminalAlive {
		mgr.Remove(sessionID)
		shutdown()
	}

	if tailCancel != nil {
		tailCancel()
		tailWG.Wait()
	}

	captured = extractAssistantTextConfigForProvider(scrollback, opts.Prompt, cfg.bannerMarkers, cfg.bannerProvider)
	tailState.Lock()
	grokSessionID = tailState.id
	streamed := tailState.streamed
	tailState.Unlock()

	if !streamed && opts.Emit != nil {
		if cfg.runnerID == "codex-tty" {
			fmt.Fprintf(opts.Stderr, "codex-tty: codex transcript not found; falling back to scrollback capture\n")
		} else if !cfg.disableTail {
			fmt.Fprintf(opts.Stderr, "grok-tty: grok session not found (updates.jsonl); falling back to scrollback capture\n")
		}
		text := strings.TrimSpace(captured)
		if text != "" {
			if emitErr := opts.Emit(types.AgentEvent{
				Type:      types.ActionMessage,
				Role:      "assistant",
				Text:      text,
				Timestamp: time.Now().UnixMilli(),
			}); emitErr != nil {
				return captured, grokSessionID, emitErr
			}
		}
	}

	if waitErr != nil {
		return captured, grokSessionID, waitErr
	}
	return captured, grokSessionID, nil
}

type runConfig struct {
	runnerID       string
	stderrPrefix   string
	registryDir    string
	bannerProvider string
	bannerMarkers  []string
	disableTail    bool
}

func (opts RunOptions) withDefaults() runConfig {
	cfg := runConfig{
		runnerID:       strings.TrimSpace(opts.RunnerID),
		stderrPrefix:   strings.TrimSpace(opts.StderrPrefix),
		registryDir:    strings.TrimSpace(opts.RegistryDir),
		bannerProvider: strings.TrimSpace(opts.BannerProvider),
		bannerMarkers:  opts.BannerMarkers,
		disableTail:    opts.DisableTail,
	}
	if cfg.runnerID == "" {
		cfg.runnerID = "grok-tty"
	}
	if cfg.stderrPrefix == "" {
		cfg.stderrPrefix = cfg.runnerID
	}
	if cfg.registryDir == "" {
		cfg.registryDir = registryDirName
	}
	if cfg.bannerProvider == "" {
		cfg.bannerProvider = "grok"
	}
	if len(cfg.bannerMarkers) == 0 {
		cfg.bannerMarkers = []string{testBannerMarker}
	}
	return cfg
}

func (cfg runConfig) buildArgv(env *agentexec.Env, settingsPath, agentPath, model, resumeSession string) ([]string, error) {
	switch cfg.runnerID {
	case "codex-tty":
		return BuildCodexCommandArgv(env, settingsPath, agentPath, model, resumeSession)
	default:
		return BuildGrokCommandArgv(env, settingsPath, agentPath, model, resumeSession)
	}
}
