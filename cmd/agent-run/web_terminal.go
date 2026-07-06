package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/agent-pro/pkgs/agentsend"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

type terminalResolution struct {
	Runner            string
	SessionID         string
	TerminalSessionID string
	Entry             *ttywatch.RegistryEntry
}

func resolveTerminal(store agentstorage.Store, runner, sessionID string) (*terminalResolution, bool) {
	if !agenttty.IsTTYRunner(runner) {
		return nil, false
	}
	sess, err := store.GetSession(runner, sessionID)
	if err != nil {
		return nil, false
	}
	terminalSessionID := strings.TrimSpace(sess.Meta.TerminalSessionID)
	if terminalSessionID == "" {
		return nil, false
	}
	resolved := &terminalResolution{
		Runner:            runner,
		SessionID:         sessionID,
		TerminalSessionID: terminalSessionID,
	}
	cfg := ttywatch.RegistryConfig{Home: store.Home(), Subdir: runner + "-registry"}
	entry, err := ttywatch.ReadRegistry(cfg, terminalSessionID)
	if err != nil || !ttywatch.TCPReachable(entry.ListenAddr) {
		return resolved, true
	}
	resolved.Entry = entry
	return resolved, true
}

func handleTerminalStatus(store agentstorage.Store, runner, sessionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resolved, ok := resolveTerminal(store, runner, sessionID)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]any{
				"available":  false,
				"runner":     runner,
				"session_id": sessionID,
			})
			return
		}
		if resolved.Entry == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"available":           false,
				"runner":              resolved.Runner,
				"session_id":          resolved.SessionID,
				"terminal_session_id": resolved.TerminalSessionID,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"available":           true,
			"runner":              resolved.Runner,
			"session_id":          resolved.SessionID,
			"terminal_session_id": resolved.TerminalSessionID,
		})
	}
}

var terminalWSUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleTerminalWebSocket(store agentstorage.Store, runner, sessionID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resolved, ok := resolveTerminal(store, runner, sessionID)
		if !ok || resolved.Entry == nil {
			http.Error(w, "terminal unavailable", http.StatusNotFound)
			return
		}

		browserConn, err := terminalWSUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer browserConn.Close()

		upstreamURL, err := ttywatch.TerminalWebSocketURL(resolved.Entry.ListenAddr, resolved.TerminalSessionID, "snapshot")
		if err != nil {
			_ = browserConn.WriteMessage(websocket.TextMessage, []byte("terminal unavailable"))
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		upstreamConn, _, err := websocket.DefaultDialer.DialContext(ctx, upstreamURL, nil)
		if err != nil {
			_ = browserConn.WriteMessage(websocket.TextMessage, []byte("terminal unavailable"))
			return
		}
		defer upstreamConn.Close()

		errCh := make(chan error, 2)
		go proxyWebSocketMessages(browserConn, upstreamConn, errCh)
		go proxyInitialTerminalMessages(upstreamConn, browserConn, errCh)
		<-errCh
	}
}

func sendPromptToLiveTerminal(store agentstorage.Store, runner, sessionID, prompt string) bool {
	resolved, ok := resolveTerminal(store, runner, sessionID)
	if !ok || resolved.Entry == nil {
		return false
	}

	provider, ok := agenttty.Get(runner)
	if !ok {
		return false
	}

	sess := agentsend.Session{
		Home:              store.Home(),
		Runner:            runner,
		TerminalSessionID: resolved.TerminalSessionID,
		ListenAddr:        resolved.Entry.ListenAddr,
	}
	if _, err := agentsend.Enqueue(store.Home(), sess, prompt); err != nil {
		return false
	}
	agentsend.StartDrainer(store.Home(), sess, provider)
	return true
}

func captureAssistantFromSnapshot(resolved *terminalResolution, prompt, runner string) string {
	if resolved == nil || resolved.Entry == nil {
		return ""
	}
	deadline := time.Now().Add(25 * time.Second)
	idleAfter := 750 * time.Millisecond
	var lastText string
	var lastActivity time.Time

	for time.Now().Before(deadline) {
		snapshotText, err := ttywatch.SnapshotText(resolved.Entry.ListenAddr, resolved.TerminalSessionID)
		if err != nil {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		text := strings.TrimSpace(agenttty.ExtractAssistantTextFromSnapshot(runner, []byte(snapshotText), prompt))
		if text != "" {
			if text != lastText {
				lastText = text
				lastActivity = time.Now()
			} else if !lastActivity.IsZero() && time.Since(lastActivity) >= idleAfter {
				return text
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	return lastText
}

func proxyInitialTerminalMessages(src, dst *websocket.Conn, errCh chan<- error) {
	mt, msg, err := src.ReadMessage()
	if err != nil {
		errCh <- err
		return
	}
	if err := dst.WriteMessage(mt, msg); err != nil {
		errCh <- err
		return
	}
	proxyWebSocketMessages(src, dst, errCh)
}

func proxyWebSocketMessages(src, dst *websocket.Conn, errCh chan<- error) {
	for {
		mt, msg, err := src.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}
		if err := dst.WriteMessage(mt, msg); err != nil {
			errCh <- err
			return
		}
	}
}