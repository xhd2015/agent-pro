package agentsync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

const (
	// Match pkgs/agenttty sessionDiscoveryGrace: 2s was too tight under parallel
	// doctest / CI load (preseed exists but filtered → context canceled).
	sessionDiscoveryGrace = 1 * time.Minute
	// Cover delayed-session fixtures (materialize ~8s after registry id) plus
	// scheduler/CI scheduling lag; 20s was too tight when sessionReady was late.
	discoveryBindTimeout = 45 * time.Second
)

func waitFindUpdatesBySessionID(ctx context.Context, grokHome, workspace, sessionID string) (string, bool) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", false
	}
	interval := 150 * time.Millisecond
	for {
		if path, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, sessionID); ok {
			return path, true
		}
		select {
		case <-ctx.Done():
			return "", false
		case <-time.After(interval):
		}
	}
}

func discoveryRunStart(opts GrokSyncOptions) time.Time {
	if !opts.SessionCreatedAt.IsZero() {
		return opts.SessionCreatedAt
	}
	return time.Now().Add(-sessionDiscoveryGrace)
}

func bootstrapGrokSession(ctx context.Context, opts GrokSyncOptions, knownGrokSessionID string) (grokSessionID, updatesPath string, err error) {
	grokHome := strings.TrimSpace(opts.GrokHome)
	if grokHome == "" {
		grokHome = agenttty.GrokHome()
	}
	workspace := strings.TrimSpace(opts.Workspace)
	if workspace == "" {
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			return "", "", wdErr
		}
		workspace = wd
	}

	prompt := strings.TrimSpace(opts.InitialPrompt)
	if prompt == "" {
		prompt = readInitialPromptFromMeta(opts.Sink.SessionDir())
	}

	discCtx, discCancel := context.WithTimeout(ctx, discoveryBindTimeout)
	defer discCancel()

	grokSessionID = strings.TrimSpace(knownGrokSessionID)
	if grokSessionID != "" {
		if path, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, grokSessionID); ok {
			updatesPath = path
			if abs, absErr := filepath.Abs(updatesPath); absErr == nil {
				updatesPath = abs
			}
			return grokSessionID, updatesPath, nil
		}
		// Only block-wait when a prompt implies a grok run is starting (e.g. composer follow-up).
		// Opening an idle seeded session must not hold grok-sync.lock waiting for updates.jsonl.
		if prompt != "" {
			if path, ok := waitFindUpdatesBySessionID(discCtx, grokHome, workspace, grokSessionID); ok {
				updatesPath = path
				if abs, absErr := filepath.Abs(updatesPath); absErr == nil {
					updatesPath = abs
				}
				return grokSessionID, updatesPath, nil
			}
			return "", "", fmt.Errorf("grok updates not found for session %s", grokSessionID)
		}
	}

	if prompt == "" {
		return "", "", nil
	}

	runStart := discoveryRunStart(opts)
	id, path, discErr := agenttty.DiscoverSession(discCtx, grokHome, workspace, prompt, runStart)
	if discErr != nil {
		return "", "", discErr
	}
	if abs, absErr := filepath.Abs(path); absErr == nil {
		path = abs
	}
	if err := opts.Sink.UpdateRunnerSessionID(id); err != nil {
		return "", "", err
	}
	return id, path, nil
}

func readInitialPromptFromMeta(sessionDir string) string {
	data, err := os.ReadFile(filepath.Join(sessionDir, "meta.json"))
	if err != nil {
		return ""
	}
	var meta struct {
		InitialPrompt string `json:"initial_prompt"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return ""
	}
	return strings.TrimSpace(meta.InitialPrompt)
}

func updatesTailStartOffset(path string, runStart time.Time) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	cutoff := runStart.Add(-sessionDiscoveryGrace)
	if info.ModTime().Before(cutoff) {
		return info.Size()
	}
	return 0
}