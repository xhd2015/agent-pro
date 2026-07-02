package agentstorage

import (
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// Store is the durable agent-run storage contract.
type Store interface {
	Home() string
	Config() (Config, error)
	SaveConfig(Config) error
	ListSessions(runner string) ([]SessionMeta, error)
	GetSession(runner, sessionID string) (*Session, error)
	CreateSession(runner, sessionID string, meta SessionMeta) error
	UpdateSessionStatus(runner, sessionID, status string) error
	UpdateSessionRunnerSessionID(runner, sessionID, runnerSessionID string) error
	UpdateSessionTerminalSessionID(runner, sessionID, terminalSessionID string) error
	AppendEvent(runner, sessionID string, ev types.AgentEvent) error
	ReadEvents(runner, sessionID string, afterOffset int64) ([]types.AgentEvent, int64, error)
	AppendMessage(runner, sessionID, text string) (Message, error)
	PopMessages(runner, sessionID string) ([]Message, error)
	ListMessages(runner, sessionID string) ([]Message, error)
}

// NewFileStore opens the default file-backed store. AGENT_RUN_HOME overrides home.
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
