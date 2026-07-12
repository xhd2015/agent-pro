package ttywatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// ReadSnapshot fetches the current screen frame for a live session via a single
// attach_mode=snapshot WebSocket (no resize, does not claim writer). Scrollback is
// left empty so rendering prefers the server screen frame.
//
// Important: do not use attach_mode=screen here. "screen" claims the writer role;
// closing that short-lived socket reaps the PTY child (stopChild on writer close),
// which breaks multi-poll waitForPrompt + inject (CSI Down / Enter / /status).
func ReadSnapshot(listenAddr, sessionID string) (frame string, scrollback string, cols, rows int, err error) {
	frame, cols, rows, err = readScreenSnapshotFrame(listenAddr, sessionID)
	if err != nil {
		return "", "", cols, rows, err
	}
	return frame, "", cols, rows, nil
}

// SnapshotText returns rendered printable snapshot text for a live session.
func SnapshotText(listenAddr, sessionID string) (string, error) {
	frame, scrollback, cols, rows, err := ReadSnapshot(listenAddr, sessionID)
	if err != nil {
		return "", err
	}
	return RenderSnapshotOutput(frame, scrollback, cols, rows), nil
}

// primeSnapshotText fetches a short-deadline printable snapshot for web attach priming.
func primeSnapshotText(listenAddr, sessionID string) (string, error) {
	cols, rows := 80, 24
	deadline := time.Now().Add(2 * time.Second)
	frame, c, r, done, err := readScreenSnapshotFrameOnce(listenAddr, sessionID, deadline, cols, rows)
	if err == nil && done && strings.TrimSpace(frame) != "" {
		cols, rows = c, r
		text := RenderSnapshotOutput(frame, "", cols, rows)
		if strings.TrimSpace(text) != "" {
			return text, nil
		}
	}
	scrollback, c, r, err := readSnapshotScrollbackWithDeadline(listenAddr, sessionID, 2*time.Second)
	if err != nil || strings.TrimSpace(scrollback) == "" {
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("empty snapshot scrollback")
	}
	if c > 0 {
		cols = c
	}
	if r > 0 {
		rows = r
	}
	text := RenderSnapshotOutput("", scrollback, cols, rows)
	if strings.TrimSpace(text) == "" {
		text = scrollback
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("empty snapshot text")
	}
	return text, nil
}

type serverMessage struct {
	Type      string `json:"type"`
	Message   string `json:"message,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

type attachRoleMessage struct {
	Type       string `json:"type"`
	AttachRole string `json:"attach_role"`
	Cols       int    `json:"cols"`
	Rows       int    `json:"rows"`
}

func terminalWebSocketURL(listenAddr, sessionID, attachMode string) (string, error) {
	base := "http://" + listenAddr
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Scheme = "ws"
	u.Path = "/api/terminal"
	q := u.Query()
	q.Set("session_id", sessionID)
	if attachMode != "" {
		q.Set("attach_mode", attachMode)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func parseServerMessage(data []byte) (handled bool, sessionID string, err error) {
	var msg serverMessage
	if json.Unmarshal(data, &msg) != nil || msg.Type == "" {
		return false, "", nil
	}
	switch msg.Type {
	case "session_id":
		return true, msg.SessionID, nil
	case "error":
		if msg.Message == "" {
			return true, "", fmt.Errorf("remote terminal error")
		}
		return true, "", fmt.Errorf("%s", msg.Message)
	default:
		return true, "", nil
	}
}

func parseAttachRoleDimensions(data []byte, cols, rows int) (int, int) {
	var msg attachRoleMessage
	if json.Unmarshal(data, &msg) != nil || msg.Type != "attach_role" {
		return cols, rows
	}
	if msg.Cols > 0 {
		cols = msg.Cols
	}
	if msg.Rows > 0 {
		rows = msg.Rows
	}
	return cols, rows
}

func readScreenSnapshotFrame(listenAddr, sessionID string) (string, int, int, error) {
	cols, rows := 80, 24
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		frame, c, r, done, err := readScreenSnapshotFrameOnce(listenAddr, sessionID, deadline, cols, rows)
		cols, rows = c, r
		if err != nil {
			return "", cols, rows, err
		}
		if done {
			return frame, cols, rows, nil
		}
	}
	return "", cols, rows, fmt.Errorf("timeout waiting for snapshot frame")
}

func readScreenSnapshotFrameOnce(listenAddr, sessionID string, deadline time.Time, cols, rows int) (frame string, outCols, outRows int, done bool, err error) {
	// Use snapshot role (not "screen"): screen claims writer and kill-on-disconnect
	// would destroy the ephemeral codex/fake TUI mid FetchStatus poll loop.
	conn, err := dialSnapshotWebSocket(listenAddr, sessionID, "snapshot")
	if err != nil {
		return "", cols, rows, false, err
	}
	defer conn.Close()

	// Snapshot servers may emit more than one binary frame before closing:
	// older ptywrap wrote an alt-screen reset prefix then raw scrollback as a
	// second message. Returning only the first frame left SnapshotText empty
	// ("\x1b[?1049l\x1b[0m"). Drain until close; prefer a true screen frame.
	var chunks [][]byte
	gotBinary := false
	for time.Now().Before(deadline) {
		remain := time.Until(deadline)
		readWait := 2 * time.Second
		if remain < readWait {
			readWait = remain
		}
		if readWait <= 0 {
			break
		}

		_ = conn.SetReadDeadline(time.Now().Add(readWait))
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				// Gorilla marks the connection corrupt after any read error; redial instead of retrying.
				if gotBinary {
					return joinSnapshotBinaryFrames(chunks), cols, rows, true, nil
				}
				return "", cols, rows, false, nil
			}
			if nerr := normalizeTerminalReadError(err); nerr != nil {
				return "", cols, rows, false, nerr
			}
			// Clean close / EOF: snapshot role closes after sendInitialFrame.
			if gotBinary {
				return joinSnapshotBinaryFrames(chunks), cols, rows, true, nil
			}
			return "", cols, rows, false, nil
		}
		switch msgType {
		case websocket.TextMessage:
			if handled, _, parseErr := parseServerMessage(data); parseErr != nil {
				return "", cols, rows, false, parseErr
			} else if handled {
				cols, rows = parseAttachRoleDimensions(data, cols, rows)
			} else if len(bytes.TrimSpace(data)) > 0 {
				// Non-control text frames: some test fakes (and legacy servers)
				// emit plain scrollback as TextMessage. Treat as content so
				// SnapshotText / sendable / exit markers work without BinaryMessage.
				gotBinary = true
				chunks = append(chunks, append([]byte(nil), data...))
			}
		case websocket.BinaryMessage:
			if len(data) == 0 {
				continue
			}
			gotBinary = true
			// Prefer a single rendered screen-snapshot frame when present.
			if isScreenSnapshotFrame(data) {
				return string(data), cols, rows, true, nil
			}
			chunks = append(chunks, append([]byte(nil), data...))
		}
	}
	if gotBinary {
		return joinSnapshotBinaryFrames(chunks), cols, rows, true, nil
	}
	return "", cols, rows, false, nil
}

// joinSnapshotBinaryFrames concatenates multi-part snapshot binary payloads.
// When any chunk is a full screen-snapshot frame, prefer the last such chunk.
func joinSnapshotBinaryFrames(chunks [][]byte) string {
	if len(chunks) == 0 {
		return ""
	}
	var lastScreen []byte
	for _, c := range chunks {
		if isScreenSnapshotFrame(c) {
			lastScreen = c
		}
	}
	if lastScreen != nil {
		return string(lastScreen)
	}
	var total int
	for _, c := range chunks {
		total += len(c)
	}
	out := make([]byte, 0, total)
	for _, c := range chunks {
		out = append(out, c...)
	}
	return string(out)
}

func readSnapshotScrollback(listenAddr, sessionID string) (string, int, int, error) {
	return readSnapshotScrollbackWithDeadline(listenAddr, sessionID, 10*time.Second)
}

func readSnapshotScrollbackWithDeadline(listenAddr, sessionID string, timeout time.Duration) (string, int, int, error) {
	cols, rows := 80, 24
	deadline := time.Now().Add(timeout)
	var out strings.Builder
	for time.Now().Before(deadline) {
		chunk, c, r, done, err := readSnapshotScrollbackOnce(listenAddr, sessionID, deadline, cols, rows)
		cols, rows = c, r
		if err != nil {
			if out.Len() > 0 {
				return out.String(), cols, rows, nil
			}
			return "", cols, rows, err
		}
		if chunk != "" {
			out.WriteString(chunk)
		}
		if done {
			return out.String(), cols, rows, nil
		}
	}
	if out.Len() > 0 {
		return out.String(), cols, rows, nil
	}
	return "", cols, rows, fmt.Errorf("timeout waiting for snapshot scrollback")
}

func readSnapshotScrollbackOnce(listenAddr, sessionID string, deadline time.Time, cols, rows int) (chunk string, outCols, outRows int, done bool, err error) {
	conn, err := dialSnapshotWebSocket(listenAddr, sessionID, "snapshot")
	if err != nil {
		return "", cols, rows, false, err
	}
	defer conn.Close()

	var buf strings.Builder
	for time.Now().Before(deadline) {
		remain := time.Until(deadline)
		readWait := 2 * time.Second
		if remain < readWait {
			readWait = remain
		}
		if readWait <= 0 {
			if buf.Len() > 0 {
				return buf.String(), cols, rows, true, nil
			}
			return "", cols, rows, false, nil
		}

		_ = conn.SetReadDeadline(time.Now().Add(readWait))
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if buf.Len() > 0 {
				return buf.String(), cols, rows, true, nil
			}
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				return "", cols, rows, false, nil
			}
			return "", cols, rows, false, normalizeTerminalReadError(err)
		}
		switch msgType {
		case websocket.TextMessage:
			if handled, _, parseErr := parseServerMessage(data); parseErr != nil {
				return "", cols, rows, false, parseErr
			} else if handled {
				cols, rows = parseAttachRoleDimensions(data, cols, rows)
			}
		case websocket.BinaryMessage:
			buf.Write(data)
		}
	}
	if buf.Len() > 0 {
		return buf.String(), cols, rows, true, nil
	}
	return "", cols, rows, false, nil
}

func dialSnapshotWebSocket(listenAddr, sessionID, attachMode string) (*websocket.Conn, error) {
	wsURL, err := terminalWebSocketURL(listenAddr, sessionID, attachMode)
	if err != nil {
		return nil, err
	}
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			snippet := strings.TrimSpace(string(body))
			if snippet != "" {
				return nil, fmt.Errorf("terminal connect failed: %s: %s", resp.Status, snippet)
			}
			return nil, fmt.Errorf("terminal connect failed: %s", resp.Status)
		}
		return nil, err
	}
	return conn, nil
}

func normalizeTerminalReadError(err error) error {
	if err == nil {
		return nil
	}
	if err == io.EOF {
		return nil
	}
	if ce, ok := err.(*websocket.CloseError); ok {
		switch ce.Code {
		case websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, 4000:
			return nil
		}
		return nil
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "broken pipe") {
		return nil
	}
	return err
}