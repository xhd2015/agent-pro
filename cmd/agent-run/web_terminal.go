package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
)

type terminalResolution struct {
	Runner            string
	SessionID         string
	TerminalSessionID string
	Entry             *ttyrunner.RegistryEntry
}

func isTTYRunner(runner string) bool {
	return ttyrunner.IsTTYRunner(runner)
}

func resolveTerminal(store agentstorage.Store, runner, sessionID string) (*terminalResolution, bool) {
	if !isTTYRunner(runner) {
		return nil, false
	}
	ttySess, err := ttyrunner.ResolveByAgentSession(store, runner, sessionID)
	if err != nil {
		return nil, false
	}
	if !ttySess.TCPReachable {
		return &terminalResolution{
			Runner:            runner,
			SessionID:         sessionID,
			TerminalSessionID: ttySess.TerminalSessionID,
		}, true
	}
	return &terminalResolution{
		Runner:            runner,
		SessionID:         sessionID,
		TerminalSessionID: ttySess.TerminalSessionID,
		Entry:             &ttySess.Registry,
	}, true
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

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		upstreamURL := terminalUpstreamWebSocketURL(resolved)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, terminalUpstreamWebSocketURL(resolved), nil)
	if err != nil {
		return false
	}
	defer conn.Close()

	input := buildTerminalPromptInput(prompt)
	if err := conn.WriteMessage(websocket.BinaryMessage, input); err != nil {
		return true
	}
	if text := captureTerminalAssistantText(conn, resolved, prompt, runner); text != "" {
		_ = store.AppendEvent(runner, sessionID, types.AgentEvent{
			Type:      types.ActionMessage,
			Role:      "assistant",
			Text:      text,
			Timestamp: time.Now().UnixMilli(),
		})
		_ = store.AppendEvent(runner, sessionID, types.AgentEvent{
			Type:      types.ActionDone,
			Timestamp: time.Now().UnixMilli(),
		})
		time.Sleep(2 * time.Second)
		_ = store.UpdateSessionStatus(runner, sessionID, "finished")
	}
	return true
}

func buildTerminalPromptInput(prompt string) []byte {
	// Ctrl+U clears stale text already typed into the interactive TUI prompt.
	return []byte("\x15" + strings.TrimSpace(prompt) + "\r")
}

func terminalUpstreamWebSocketURL(resolved *terminalResolution) string {
	u := url.URL{
		Scheme: "ws",
		Host:   resolved.Entry.ListenAddr,
		Path:   "/api/terminal",
	}
	q := u.Query()
	q.Set("session_id", resolved.TerminalSessionID)
	q.Set("attach_mode", "snapshot")
	u.RawQuery = q.Encode()
	return u.String()
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

func captureTerminalAssistantText(conn *websocket.Conn, resolved *terminalResolution, prompt, runner string) string {
	deadline := time.Now().Add(1 * time.Second)
	idleAfter := 500 * time.Millisecond
	var lastActivity time.Time
	var out strings.Builder
	for {
		readDeadline := deadline
		if !lastActivity.IsZero() {
			idleDeadline := lastActivity.Add(idleAfter)
			if idleDeadline.Before(readDeadline) {
				readDeadline = idleDeadline
			}
		}
		_ = conn.SetReadDeadline(readDeadline)
		_, msg, err := conn.ReadMessage()
		_ = conn.SetReadDeadline(time.Time{})
		if err != nil {
			break
		}
		if len(msg) == 0 {
			continue
		}
		out.Write(msg)
		lastActivity = time.Now()
	}
	if text := extractLiveTerminalAssistantText(out.String(), prompt, runner); text != "" {
		return text
	}
	return pollTerminalAssistantSnapshot(resolved, prompt, runner, 25*time.Second)
}

var (
	terminalOSCRe                 = regexp.MustCompile(`\x1b\][^\x07]*(?:\x07|\x1b\\)`)
	terminalANSIEscapeRe          = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)
	terminalControlJSONFragmentRe = regexp.MustCompile(`\{[^{}\n]*(?:"type"\s*:\s*"session_id"|"terminal_session_id"|"status"\s*:\s*"running"|"session_id")[^{}\n]*\}`)
)

func extractLiveTerminalAssistantText(raw, prompt, runner string) string {
	plain := terminalOSCRe.ReplaceAllString(raw, "")
	plain = terminalANSIEscapeRe.ReplaceAllString(plain, "")
	plain = terminalControlJSONFragmentRe.ReplaceAllString(plain, "")
	plain = strings.ReplaceAll(plain, "\r", "\n")
	plain = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, plain)
	prompt = strings.TrimSpace(prompt)
	if response := extractGlobalTerminalResponseMarker(plain); response != "" {
		return response
	}
	if isCodexLiveRunner(runner) {
		if text := extractCodexLiveAssistantText(plain, prompt); text != "" {
			return text
		}
	}
	candidate := plain
	if prompt != "" {
		compactPlain := strings.ToLower(plain)
		compactPrompt := strings.ToLower(prompt)
		if idx := strings.LastIndex(compactPlain, compactPrompt); idx >= 0 {
			candidate = plain[idx+len(prompt):]
		}
	}

	var kept []string
	for _, line := range strings.Split(candidate, "\n") {
		line = cleanLiveTerminalLine(line, prompt, runner)
		if line != "" {
			kept = append(kept, line)
		}
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func isCodexLiveRunner(runner string) bool {
	return registry.AgentRunnerID(strings.TrimSpace(runner)) == registry.AgentRunnerCodexTTY
}

func extractCodexLiveAssistantText(plain, prompt string) string {
	prompt = strings.TrimSpace(prompt)
	var kept []string
	for _, line := range strings.Split(plain, "\n") {
		line = cleanLiveTerminalTextLine(line)
		if line == "" {
			continue
		}
		if bullet := extractCodexLiveBulletText(line, prompt); bullet != "" {
			kept = append(kept, bullet)
			continue
		}
		line = trimCodexPromptSegments(line)
		line = cleanLiveTerminalTextLine(line)
		if line == "" || skipCodexLiveLine(line, prompt) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func extractCodexLiveBulletText(line, prompt string) string {
	if !strings.Contains(line, "•") {
		return ""
	}
	var kept []string
	for _, segment := range strings.Split(line, "•")[1:] {
		if idx := strings.Index(segment, "›"); idx >= 0 {
			segment = segment[:idx]
		}
		segment = cleanLiveTerminalTextLine(segment)
		if segment == "" || skipCodexLiveLine(segment, prompt) {
			continue
		}
		lower := strings.ToLower(segment)
		if strings.HasPrefix(lower, "working") ||
			strings.HasPrefix(lower, "running ") ||
			strings.HasPrefix(lower, "starting ") ||
			strings.HasPrefix(lower, "queued ") ||
			strings.Contains(lower, "esc to interrupt") {
			continue
		}
		kept = append(kept, segment)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

func trimCodexPromptSegments(line string) string {
	for strings.Contains(line, "›") {
		idx := strings.Index(line, "›")
		nextBullet := strings.Index(line[idx:], "•")
		if nextBullet < 0 {
			return strings.TrimSpace(line[:idx])
		}
		line = strings.TrimSpace(line[idx+nextBullet:])
	}
	return line
}

func cleanLiveTerminalTextLine(line string) string {
	line = strings.TrimSpace(line)
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' {
			return -1
		}
		return r
	}, line)
}

func skipCodexLiveLine(line, prompt string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return true
	}
	for _, marker := range []string{"CODEX_TTY_BANNER", "[Terminal exited]"} {
		if strings.Contains(line, marker) {
			return true
		}
	}
	if isTerminalControlLine(line) || terminalControlJSONFragmentRe.MatchString(line) {
		return true
	}
	if prompt != "" && strings.EqualFold(line, strings.TrimSpace(prompt)) {
		return true
	}
	if strings.Contains(line, "›") ||
		strings.ContainsAny(line, "╭╮╰╯│─") ||
		strings.Contains(line, ">4;0m") ||
		strings.Contains(line, ">7u") {
		return true
	}
	lower := strings.ToLower(line)
	compact := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "", "`", "", `"`, "", "'", "").Replace(lower)
	if lower == "codex" ||
		lower == "grok" ||
		strings.Contains(lower, "openai codex") ||
		strings.Contains(lower, "[features].codex_hooks") ||
		strings.Contains(lower, "[features].hooks") ||
		strings.Contains(lower, "developers.openai.com/codex") ||
		strings.HasPrefix(lower, "enable it with") ||
		strings.HasPrefix(lower, "for details") ||
		strings.HasPrefix(lower, "tip:") ||
		strings.HasPrefix(lower, "permissions:") ||
		strings.HasPrefix(lower, "model:") ||
		strings.HasPrefix(lower, "directory:") ||
		strings.Contains(compact, "model:loading") ||
		strings.Contains(lower, "starting mcp servers") ||
		strings.Contains(lower, "booting mcp") ||
		strings.Contains(lower, "running stop hook") ||
		strings.Contains(lower, "running userpromptsubmit hook") ||
		strings.HasPrefix(lower, "working") {
		return true
	}
	return false
}

func cleanLiveTerminalLine(line, prompt, runner string) string {
	line = strings.TrimSpace(line)
	line = terminalControlJSONFragmentRe.ReplaceAllString(line, "")
	line = strings.TrimPrefix(line, "echo:")
	line = strings.TrimSpace(line)
	for _, marker := range []string{"CODEX_TTY_BANNER", "GROK_TTY_BANNER"} {
		line = strings.ReplaceAll(line, marker, "")
	}
	for _, leader := range []string{"Codex ›", "Codex ›", "Grok ›", "Grok ›"} {
		if strings.Contains(line, leader) {
			parts := strings.Split(line, leader)
			line = strings.TrimSpace(parts[len(parts)-1])
		}
	}
	if prompt != "" && strings.EqualFold(strings.TrimSpace(line), prompt) {
		return ""
	}
	if prompt != "" && strings.Contains(strings.ToLower(line), strings.ToLower(prompt)) {
		if strings.Contains(line, "›") || strings.HasPrefix(line, "echo:") {
			return ""
		}
	}
	if strings.Contains(line, "[Terminal exited]") {
		return ""
	}
	if isTerminalControlLine(line) {
		return ""
	}
	if strings.TrimSpace(line) == "" {
		return ""
	}
	return line
}

func isTerminalControlLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" || !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
		return false
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		return false
	}
	if _, ok := obj["terminal_session_id"]; ok {
		return true
	}
	if typ, ok := obj["type"].(string); ok {
		switch typ {
		case "session_id", "resize", "status":
			return true
		}
	}
	_, hasSessionID := obj["session_id"]
	_, hasStatus := obj["status"]
	return hasSessionID && hasStatus
}

func extractGlobalTerminalResponseMarker(plain string) string {
	var marked []string
	for _, line := range strings.Split(plain, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch {
		case strings.Contains(line, "FOLLOWUP_RESPONSE"):
			marked = append(marked, line)
		case strings.Contains(line, "SUBMITTED:"):
			marked = append(marked, line)
		}
	}
	if len(marked) == 0 {
		return ""
	}
	return strings.TrimSpace(marked[len(marked)-1])
}

func pollTerminalAssistantSnapshot(resolved *terminalResolution, prompt, runner string, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snapshot := readTerminalSnapshot(resolved)
		if text := extractLiveTerminalAssistantText(snapshot, prompt, runner); text != "" {
			return text
		}
		time.Sleep(250 * time.Millisecond)
	}
	return ""
}

func readTerminalSnapshot(resolved *terminalResolution) string {
	if resolved == nil || resolved.Entry == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, terminalUpstreamWebSocketURL(resolved), nil)
	if err != nil {
		return ""
	}
	defer conn.Close()

	var out strings.Builder
	for {
		_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, msg, err := conn.ReadMessage()
		_ = conn.SetReadDeadline(time.Time{})
		if err != nil {
			break
		}
		out.Write(msg)
	}
	return out.String()
}
