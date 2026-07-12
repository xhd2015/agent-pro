package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

var terminalWSUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleTerminalStatus(store agentstorage.Store, runner, sessionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ttySess, err := agenttty.ResolveTerminalStatus(store, runner, sessionID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"available":  false,
				"runner":     runner,
				"session_id": sessionID,
			})
			return
		}
		if ttySess.Meta != nil && strings.TrimSpace(ttySess.Meta.TerminalSessionID) == "" && ttySess.TerminalSessionID != "" {
			_ = store.UpdateSessionTerminalSessionID(sessionID, ttySess.TerminalSessionID)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"available":           ttySess.TCPReachable,
			"runner":              ttySess.RunnerID,
			"session_id":          ttySess.AgentSessionID,
			"terminal_session_id": ttySess.TerminalSessionID,
		})
	}
}

func handleTerminalWebSocket(store agentstorage.Store, runner, sessionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ttySess, err := agenttty.ResolveByAgentSession(store, runner, sessionID)
		if err != nil || !ttySess.TCPReachable {
			http.Error(w, "terminal unavailable", http.StatusNotFound)
			return
		}

		browserConn, err := terminalWSUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer browserConn.Close()

		sink := ttywatch.NewWebSocketAttachSink(browserConn)
		cfg := ttywatch.AttachRelayConfig{
			ExitOnTerminalExit:           false,
			SkipScreenSnapshotConversion: true,
			Cols:                         80,
			Rows:                         24,
			OnConnect: ttywatch.WebAttachOnConnect(
				ttySess.Registry.ListenAddr,
				ttySess.TerminalSessionID,
				sink.OutputWriter(),
				80,
				24,
			),
		}
		_ = ttywatch.AttachRelay(r.Context(), ttySess.Registry.ListenAddr, ttySess.TerminalSessionID, "attach", cfg, sink)
	}
}