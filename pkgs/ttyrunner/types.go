package ttyrunner

import (
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/groktty"
)

// RegistryEntry is the live TTY process index entry (alias of groktty.RegistryEntry).
type RegistryEntry = groktty.RegistryEntry

// WritableStatus reports whether a TTY session is ready to receive injected input.
type WritableStatus struct {
	Ready  bool
	Reason string
	State  string
}

// TTYSnapshot is the denormalized TTY cross-ref written to sessions/.../tty.json.
type TTYSnapshot struct {
	RunnerID          string `json:"runner_id"`
	AgentSessionID    string `json:"agent_session_id"`
	TerminalSessionID string `json:"terminal_session_id"`
	ListenAddr        string `json:"listen_addr"`
	PID               int    `json:"pid"`
	CreatedAt         string `json:"created_at"`
	ScreenStatus      string `json:"screen_status,omitempty"`
	Alive             bool   `json:"alive"`
}

// TTYSession is the unified resolver result for status, attach, and send.
type TTYSession struct {
	RunnerID          string
	AgentSessionID    string
	TerminalSessionID string
	Registry          RegistryEntry
	Meta              *agentstorage.SessionMeta
	TTY               *TTYSnapshot
	TCPReachable      bool
	ScreenStatus      string
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}