package agentsync

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// ReconcileOptions configures a reconcile pass.
type ReconcileOptions struct {
	Home      string
	GrokHome  string
	Runner    string
	SessionID string
}

type reconcileSessionMeta struct {
	Runner          string `json:"runner"`
	SessionID       string `json:"session_id"`
	RunnerSessionID string `json:"runner_session_id"`
	InitialPrompt   string `json:"initial_prompt"`
	Status          string `json:"status"`
	Workspace       string `json:"workspace"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// ReconcileOnce attempts to heal a single session or scan candidates when SessionID is empty.
func ReconcileOnce(ctx context.Context, opts ReconcileOptions) error {
	runner := strings.TrimSpace(opts.Runner)
	if runner == "" {
		runner = "grok-tty"
	}
	if strings.TrimSpace(opts.SessionID) != "" {
		return reconcileSession(ctx, opts, runner, opts.SessionID)
	}
	return scanAndReconcile(ctx, opts, runner)
}

// StartReconciler runs reconcile passes on a fixed interval until ctx is cancelled.
func StartReconciler(ctx context.Context, interval time.Duration, opts ReconcileOptions) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = ReconcileOnce(ctx, opts)
			}
		}
	}()
}

func scanAndReconcile(ctx context.Context, opts ReconcileOptions, runner string) error {
	home := strings.TrimSpace(opts.Home)
	if home == "" {
		return nil
	}
	root := filepath.Join(home, "sessions", runner)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		if err := reconcileSession(ctx, opts, runner, ent.Name()); err != nil {
			return err
		}
	}
	return nil
}

func reconcileSession(ctx context.Context, opts ReconcileOptions, runner, sessionID string) error {
	if GrokSyncWorkerActive(runner, sessionID) {
		return nil
	}
	meta, sessionDir, err := loadReconcileMeta(opts.Home, runner, sessionID)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !shouldReconcileSession(opts.Home, runner, sessionID, meta, sessionDir) {
		return nil
	}

	grokHome := strings.TrimSpace(opts.GrokHome)
	if grokHome == "" {
		grokHome = agenttty.GrokHome()
	}
	workspace := reconcileWorkspace(opts.Home, runner, sessionID, meta.Workspace)
	grokSessionID := strings.TrimSpace(meta.RunnerSessionID)
	updatesPath := ""
	if grokSessionID != "" {
		if path, ok := agenttty.FindUpdatesBySessionID(grokHome, workspace, grokSessionID); ok {
			updatesPath = path
		}
	}

	sink := NewFileGrokSyncSink(sessionDir, grokSessionID, updatesPath)
	return EnsureGrokSync(ctx, GrokSyncOptions{
		Runner:           runner,
		SessionID:        sessionID,
		GrokSessionID:    grokSessionID,
		UpdatesPath:      updatesPath,
		Workspace:        workspace,
		GrokHome:         grokHome,
		InitialPrompt:    meta.InitialPrompt,
		SessionCreatedAt: parseRFC3339(meta.CreatedAt),
		Sink:             sink,
	})
}

func loadReconcileMeta(home, runner, sessionID string) (reconcileSessionMeta, string, error) {
	sessionDir := filepath.Join(home, "sessions", runner, sessionID)
	data, err := os.ReadFile(filepath.Join(sessionDir, "meta.json"))
	if err != nil {
		return reconcileSessionMeta{}, sessionDir, err
	}
	var meta reconcileSessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return reconcileSessionMeta{}, sessionDir, err
	}
	return meta, sessionDir, nil
}

func shouldReconcileSession(home, runner, sessionID string, meta reconcileSessionMeta, sessionDir string) bool {
	if !sessionMetaUpdatedWithin(meta, 24*time.Hour) {
		return false
	}
	if meta.Status == "running" {
		return true
	}
	if ttyAlive(home, runner, sessionID) {
		return true
	}
	if eventsEmpty(sessionDir) && strings.TrimSpace(meta.InitialPrompt) != "" {
		return true
	}
	if checkpointBehindEOF(sessionDir) {
		return true
	}
	return false
}

func sessionMetaUpdatedWithin(meta reconcileSessionMeta, within time.Duration) bool {
	updated := parseRFC3339(meta.UpdatedAt)
	if updated.IsZero() {
		updated = parseRFC3339(meta.CreatedAt)
	}
	if updated.IsZero() {
		return true
	}
	return time.Since(updated) <= within
}

func parseRFC3339(value string) time.Time {
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

func reconcileWorkspace(home, runner, sessionID, metaWorkspace string) string {
	ws := strings.TrimSpace(metaWorkspace)
	if ws != "" {
		sessSuffix := filepath.Join("sessions", runner, sessionID)
		if !strings.Contains(ws, sessSuffix) {
			return ws
		}
	}
	home = strings.TrimSpace(home)
	if home != "" {
		return filepath.Dir(home)
	}
	return ws
}

func eventsEmpty(sessionDir string) bool {
	info, err := os.Stat(filepath.Join(sessionDir, "events.jsonl"))
	if err != nil {
		return true
	}
	return info.Size() == 0
}

func checkpointBehindEOF(sessionDir string) bool {
	cpPath := filepath.Join(sessionDir, grokSyncCheckpointFile)
	cpData, err := os.ReadFile(cpPath)
	if err != nil {
		return false
	}
	var cp GrokSyncState
	if err := json.Unmarshal(cpData, &cp); err != nil {
		return false
	}
	if strings.TrimSpace(cp.UpdatesPath) == "" {
		return false
	}
	info, err := os.Stat(cp.UpdatesPath)
	if err != nil {
		return false
	}
	return cp.UpdatesOffset < info.Size()
}

func ttyAlive(home, runner, sessionID string) bool {
	data, err := os.ReadFile(filepath.Join(home, "sessions", runner, sessionID, "tty.json"))
	if err != nil {
		return false
	}
	var snap struct {
		Alive bool `json:"alive"`
	}
	if err := json.Unmarshal(data, &snap); err != nil {
		return false
	}
	return snap.Alive
}