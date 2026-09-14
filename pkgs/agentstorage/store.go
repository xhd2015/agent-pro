package agentstorage

import (
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// Store is the durable agent-run storage contract.
// Session identity is bare sessionID; runner is metadata on SessionMeta only.
type Store interface {
	Home() string
	Config() (Config, error)
	SaveConfig(Config) error
	ListSessions() ([]SessionMeta, error)
	ClearAllSessions() error
	GetSession(sessionID string) (*Session, error)
	CreateSession(sessionID string, meta SessionMeta) error
	UpdateSessionStatus(sessionID, status string) error
	UpdateSessionRunnerSessionID(sessionID, runnerSessionID string) error
	// ClearSessionRunnerSessionID removes the live runner_session_id and the
	// current family's RunnerSessions slot so the next AutoSendOrResume can
	// ModeRun that family (orphan/missing provider sessions). Other families
	// in the map are kept. No-op when the current family is already unbound.
	ClearSessionRunnerSessionID(sessionID string) error
	// ActivateSessionRunner sets meta.runner to runner and restores the live
	// runner_session_id from RunnerSessions[family] when present (else clears
	// the live slot). Other families in the map are kept.
	ActivateSessionRunner(sessionID, runner string) error
	UpdateSessionTerminalSessionID(sessionID, terminalSessionID string) error
	// UpdateSessionWorkspace sets meta.workspace (e.g. after Grok session relocate).
	UpdateSessionWorkspace(sessionID, workspace string) error
	// UpdateSessionEnvConfig writes session-scoped TTY child env fields.
	// prependPaths/env replace the stored lists; configHome replaces when non-empty.
	UpdateSessionEnvConfig(sessionID string, prependPaths, env []string, configHome string) error
	AppendEvent(sessionID string, ev types.AgentEvent) error
	ReadEvents(sessionID string, afterOffset int64) ([]types.AgentEvent, int64, error)
	AppendMessage(sessionID, text string) (Message, error)
	PopMessages(sessionID string) ([]Message, error)
	ListMessages(sessionID string) ([]Message, error)
}

// NewFileStore opens the file-backed store at home. Non-empty home wins;
// empty home falls back to AGENT_RUN_HOME or ~/.agent-run.
func NewFileStore(home string) (Store, error) {
	resolved, err := resolveHome(home)
	if err != nil {
		return nil, err
	}
	if err := osMkdirAll(resolved); err != nil {
		return nil, err
	}
	return &fileStore{home: resolved}, nil
}
