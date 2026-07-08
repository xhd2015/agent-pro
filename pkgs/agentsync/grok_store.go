package agentsync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

// NewFileGrokSyncSink writes events.jsonl and grok-sync.json under sessionDir.
func NewFileGrokSyncSink(sessionDir, grokSessionID, updatesPath string) GrokSyncSink {
	return &fileGrokSyncSink{
		sessionDir:    sessionDir,
		grokSessionID: grokSessionID,
		updatesPath:   updatesPath,
	}
}

type fileGrokSyncSink struct {
	sessionDir    string
	grokSessionID string
	updatesPath   string
	mu            sync.Mutex
}

func (s *fileGrokSyncSink) SessionDir() string {
	return s.sessionDir
}

func (s *fileGrokSyncSink) eventsPath() string {
	return filepath.Join(s.sessionDir, "events.jsonl")
}

func (s *fileGrokSyncSink) checkpointPath() string {
	return filepath.Join(s.sessionDir, grokSyncCheckpointFile)
}

func (s *fileGrokSyncSink) metaPath() string {
	return filepath.Join(s.sessionDir, "meta.json")
}

func (s *fileGrokSyncSink) AppendEvent(ev types.AgentEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.sessionDir, 0755); err != nil {
		return err
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.eventsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s\n", line)
	return err
}

func (s *fileGrokSyncSink) LoadCheckpoint() (GrokSyncState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.checkpointPath())
	if err != nil {
		if os.IsNotExist(err) {
			return GrokSyncState{}, nil
		}
		return GrokSyncState{}, err
	}
	var st GrokSyncState
	if err := json.Unmarshal(data, &st); err != nil {
		return GrokSyncState{}, err
	}
	return st, nil
}

func (s *fileGrokSyncSink) SaveCheckpoint(st GrokSyncState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.sessionDir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return os.WriteFile(s.checkpointPath(), data, 0644)
}

func (s *fileGrokSyncSink) OnTurnCompleted() error {
	return nil
}

func (s *fileGrokSyncSink) UpdateRunnerSessionID(runnerSessionID string) error {
	runnerSessionID = strings.TrimSpace(runnerSessionID)
	if runnerSessionID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.metaPath())
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		data = []byte("{}")
	}
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		return err
	}
	if meta == nil {
		meta = map[string]any{}
	}
	meta["runner_session_id"] = runnerSessionID
	encoded, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.sessionDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(s.metaPath(), encoded, 0644)
}

// StoreGrokSyncSink persists grok sync output via agentstorage.Store.
type StoreGrokSyncSink struct {
	store         agentstorage.Store
	runner        string
	sessionID     string
	grokSessionID string
	updatesPath   string
}

// NewStoreGrokSyncSink creates a store-backed GrokSyncSink for a session.
func NewStoreGrokSyncSink(store agentstorage.Store, runner, sessionID, grokSessionID, updatesPath string) *StoreGrokSyncSink {
	return &StoreGrokSyncSink{
		store:         store,
		runner:        runner,
		sessionID:     sessionID,
		grokSessionID: grokSessionID,
		updatesPath:   updatesPath,
	}
}

func (s *StoreGrokSyncSink) SessionDir() string {
	return filepath.Join(s.store.Home(), "sessions", s.runner, s.sessionID)
}

func normalizeGrokUserPromptText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func (s *StoreGrokSyncSink) checkpointPath() string {
	return filepath.Join(s.SessionDir(), grokSyncCheckpointFile)
}

func (s *StoreGrokSyncSink) AppendEvent(ev types.AgentEvent) error {
	if ev.Type == types.ActionMessage && strings.TrimSpace(ev.Role) == "" {
		ev.Role = "assistant"
	}
	if ev.Type == types.ActionMessage && ev.Timestamp == 0 {
		ev.Timestamp = time.Now().UnixMilli()
	}
	if ev.Type == types.ActionMessage && ev.Role == "user" {
		text := normalizeGrokUserPromptText(ev.Text)
		if text != "" {
			events, _, err := s.store.ReadEvents(s.runner, s.sessionID, 0)
			if err == nil {
				for _, existing := range events {
					if existing.Type != types.ActionMessage || existing.Role != "user" {
						continue
					}
					if normalizeGrokUserPromptText(existing.Text) == text {
						return nil
					}
				}
			}
		}
	}
	return s.store.AppendEvent(s.runner, s.sessionID, ev)
}

func (s *StoreGrokSyncSink) LoadCheckpoint() (GrokSyncState, error) {
	data, err := os.ReadFile(s.checkpointPath())
	if err != nil {
		if os.IsNotExist(err) {
			return GrokSyncState{}, nil
		}
		return GrokSyncState{}, err
	}
	var st GrokSyncState
	if err := json.Unmarshal(data, &st); err != nil {
		return GrokSyncState{}, err
	}
	return st, nil
}

func (s *StoreGrokSyncSink) SaveCheckpoint(st GrokSyncState) error {
	if err := os.MkdirAll(s.SessionDir(), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return os.WriteFile(s.checkpointPath(), data, 0644)
}

func (s *StoreGrokSyncSink) OnTurnCompleted() error {
	return s.store.UpdateSessionStatus(s.runner, s.sessionID, "finished")
}

func (s *StoreGrokSyncSink) UpdateRunnerSessionID(runnerSessionID string) error {
	return s.store.UpdateSessionRunnerSessionID(s.runner, s.sessionID, runnerSessionID)
}

// runGrokSyncSink optionally forwards emitted events to a callback while persisting checkpoints.
type runGrokSyncSink struct {
	emit    func(types.AgentEvent) error
	checkpt *StoreGrokSyncSink
}

// NewRunGrokSyncSink wraps a store sink and optionally forwards emitted events.
func NewRunGrokSyncSink(emit func(types.AgentEvent) error, checkpt *StoreGrokSyncSink) GrokSyncSink {
	return &runGrokSyncSink{emit: emit, checkpt: checkpt}
}

func (s *runGrokSyncSink) SessionDir() string {
	return s.checkpt.SessionDir()
}

func (s *runGrokSyncSink) AppendEvent(ev types.AgentEvent) error {
	if s.emit == nil {
		return s.checkpt.AppendEvent(ev)
	}
	return s.emit(ev)
}

func (s *runGrokSyncSink) LoadCheckpoint() (GrokSyncState, error) {
	return s.checkpt.LoadCheckpoint()
}

func (s *runGrokSyncSink) SaveCheckpoint(st GrokSyncState) error {
	return s.checkpt.SaveCheckpoint(st)
}

func (s *runGrokSyncSink) OnTurnCompleted() error {
	return s.checkpt.OnTurnCompleted()
}

func (s *runGrokSyncSink) UpdateRunnerSessionID(runnerSessionID string) error {
	return s.checkpt.UpdateRunnerSessionID(runnerSessionID)
}