package agentsync

import (
	"context"
	"fmt"
	"sync"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

const grokSyncCheckpointFile = "grok-sync.json"

// GrokSyncState is the persisted grok updates tail checkpoint.
type GrokSyncState struct {
	GrokSessionID string `json:"grok_session_id"`
	UpdatesPath   string `json:"updates_path"`
	UpdatesOffset int64  `json:"updates_offset"`
	TurnIndex     int    `json:"turn_index"`
}

// GrokSyncSink persists grok sync worker output and checkpoints.
type GrokSyncSink interface {
	SessionDir() string
	AppendEvent(ev types.AgentEvent) error
	LoadCheckpoint() (GrokSyncState, error)
	SaveCheckpoint(GrokSyncState) error
	OnTurnCompleted() error
	UpdateRunnerSessionID(runnerSessionID string) error
}

// GrokSyncOptions configures a persistent grok sync worker.
type GrokSyncOptions struct {
	GrokHome         string
	InitialPrompt    string
	SessionCreatedAt time.Time
	Workspace        string
	Runner           string
	SessionID        string
	GrokSessionID    string
	UpdatesPath      string
	Sink             GrokSyncSink
}

type grokSyncWorker struct {
	cancel      context.CancelFunc
	done        chan struct{}
	releaseLock func()
}

var (
	grokSyncRegistryMu sync.Mutex
	grokSyncWorkers    = map[string]*grokSyncWorker{}
)

func grokSyncWorkerKey(runner, sessionID string) string {
	return runner + "\x00" + sessionID
}

// EnsureGrokSync starts (or reuses) a single persistent grok sync worker per session.
func EnsureGrokSync(ctx context.Context, opts GrokSyncOptions) error {
	_ = ctx
	if opts.Sink == nil {
		return fmt.Errorf("grok sync sink is required")
	}
	key := grokSyncWorkerKey(opts.Runner, opts.SessionID)

	grokSyncRegistryMu.Lock()
	if _, ok := grokSyncWorkers[key]; ok {
		grokSyncRegistryMu.Unlock()
		return nil
	}

	release, acquired, err := tryAcquireSessionLock(opts.Sink.SessionDir())
	if err != nil {
		grokSyncRegistryMu.Unlock()
		return err
	}
	if !acquired {
		grokSyncRegistryMu.Unlock()
		return nil
	}

	workerCtx, cancel := context.WithCancel(context.Background())
	w := &grokSyncWorker{
		cancel:      cancel,
		done:        make(chan struct{}),
		releaseLock: release,
	}
	grokSyncWorkers[key] = w
	grokSyncRegistryMu.Unlock()

	go func() {
		defer close(w.done)
		defer w.releaseLock()
		defer func() {
			grokSyncRegistryMu.Lock()
			delete(grokSyncWorkers, key)
			grokSyncRegistryMu.Unlock()
		}()
		_ = runGrokSyncWorker(workerCtx, opts)
	}()
	return nil
}

// StopGrokSync stops the grok sync worker for a session, if running.
func StopGrokSync(runner, sessionID string) error {
	key := grokSyncWorkerKey(runner, sessionID)
	grokSyncRegistryMu.Lock()
	w, ok := grokSyncWorkers[key]
	grokSyncRegistryMu.Unlock()
	if !ok {
		return nil
	}
	w.cancel()
	<-w.done
	return nil
}

// GrokSyncWorkerCount returns the number of active grok sync workers.
func GrokSyncWorkerCount() int {
	grokSyncRegistryMu.Lock()
	defer grokSyncRegistryMu.Unlock()
	return len(grokSyncWorkers)
}

// GrokSyncWorkerActive reports whether a grok sync worker is running for the session.
// It checks the in-process registry and, when absent, whether another process holds
// the per-session grok-sync.lock (web server subprocess integration tests).
func GrokSyncWorkerActive(runner, sessionID string) bool {
	grokSyncRegistryMu.Lock()
	_, ok := grokSyncWorkers[grokSyncWorkerKey(runner, sessionID)]
	grokSyncRegistryMu.Unlock()
	if ok {
		return true
	}
	sessionDir, found := resolveGrokSyncSessionDir(runner, sessionID)
	if !found {
		return false
	}
	return grokSyncLockHeld(sessionDir)
}