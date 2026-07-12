package agentui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agentsync"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func resolveGrokSessionID(store agentstorage.Store, runner, sessionID string) string {
	if sess, err := store.GetSession(sessionID); err == nil {
		if id := strings.TrimSpace(sess.Meta.RunnerSessionID); id != "" {
			return id
		}
	}
	return strings.TrimSpace(os.Getenv("AGENT_RUN_GROK_TTY_GROK_SESSION_ID"))
}

func sessionCreatedAt(store agentstorage.Store, runner, sessionID string) time.Time {
	sess, err := store.GetSession(sessionID)
	if err != nil {
		return time.Now().Add(-2 * time.Second)
	}
	if t := parseSessionTime(sess.Meta.CreatedAt); !t.IsZero() {
		return t
	}
	return time.Now().Add(-2 * time.Second)
}

func parseSessionTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Time{}
}

func initialPromptForSession(store agentstorage.Store, runner, sessionID, fallback string) string {
	if sess, err := store.GetSession(sessionID); err == nil {
		if p := strings.TrimSpace(sess.Meta.InitialPrompt); p != "" {
			return p
		}
	}
	return strings.TrimSpace(fallback)
}

func ensureGrokSyncForSession(ctx context.Context, opts RunOptions, grokSessionID string, emit func(types.AgentEvent) error) error {
	if opts.Runner != "grok-tty" || opts.Store == nil {
		return nil
	}
	workspace := opts.Workspace
	if workspace == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		workspace = wd
	}
	grokHome := agenttty.GrokHomeForRunner(opts.AgentRunnerConfigHome)
	grokSessionID = strings.TrimSpace(grokSessionID)
	if grokSessionID == "" {
		grokSessionID = resolveGrokSessionID(opts.Store, opts.Runner, opts.SessionID)
	}

	updatesPath := ""
	if grokSessionID != "" {
		if path, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, grokSessionID); ok {
			updatesPath = path
			if abs, err := filepath.Abs(updatesPath); err == nil {
				updatesPath = abs
			}
		}
	}

	checkpt := agentsync.NewStoreGrokSyncSink(opts.Store, opts.Runner, opts.SessionID, grokSessionID, updatesPath)
	var sink agentsync.GrokSyncSink = checkpt
	if emit != nil {
		sink = agentsync.NewRunGrokSyncSink(emit, checkpt)
	}
	return agentsync.EnsureGrokSync(ctx, agentsync.GrokSyncOptions{
		Runner:           opts.Runner,
		SessionID:        opts.SessionID,
		GrokSessionID:    grokSessionID,
		UpdatesPath:      updatesPath,
		Workspace:        workspace,
		GrokHome:         grokHome,
		InitialPrompt:    initialPromptForSession(opts.Store, opts.Runner, opts.SessionID, opts.Prompt),
		SessionCreatedAt: sessionCreatedAt(opts.Store, opts.Runner, opts.SessionID),
		Sink:             sink,
	})
}

func startGrokSyncPoller(ctx context.Context, opts RunOptions, emit func(types.AgentEvent) error) {
	if opts.Runner != "grok-tty" || !opts.KeepTerminalAlive || opts.Store == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := ensureGrokSyncForSession(context.Background(), opts, "", emit); err != nil {
					continue
				}
				if agentsync.GrokSyncWorkerActive(opts.Runner, opts.SessionID) {
					return
				}
			}
		}
	}()
}