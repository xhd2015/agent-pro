package ttywatch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// ReadSnapshot fetches the current screen frame and scrollback for a live session.
func ReadSnapshot(listenAddr, sessionID string) (frame string, scrollback string, cols, rows int, err error) {
	frame, cols, rows, err = readScreenSnapshotFrame(listenAddr, sessionID)
	if err != nil {
		return "", "", cols, rows, err
	}
	scrollback, _, _, scrollErr := readSnapshotScrollback(listenAddr, sessionID)
	if scrollErr != nil {
		scrollback = ""
	}
	return frame, scrollback, cols, rows, nil
}

// SnapshotText returns rendered printable snapshot text for a live session.
func SnapshotText(listenAddr, sessionID string) (string, error) {
	frame, scrollback, cols, rows, err := ReadSnapshot(listenAddr, sessionID)
	if err != nil {
		return "", err
	}
	return RenderSnapshotOutput(frame, scrollback, cols, rows), nil
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
	conn, err := dialSnapshotWebSocket(listenAddr, sessionID, "screen")
	if err != nil {
		return "", 0, 0, err
	}
	defer conn.Close()

	cols, rows := 80, 24
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				continue
			}
			return "", cols, rows, err
		}
		switch msgType {
		case websocket.TextMessage:
			if handled, _, parseErr := parseServerMessage(data); parseErr != nil {
				return "", cols, rows, parseErr
			} else if handled {
				cols, rows = parseAttachRoleDimensions(data, cols, rows)
			}
		case websocket.BinaryMessage:
			return string(data), cols, rows, nil
		}
	}
	return "", cols, rows, fmt.Errorf("timeout waiting for snapshot frame")
}

func readSnapshotScrollback(listenAddr, sessionID string) (string, int, int, error) {
	conn, err := dialSnapshotWebSocket(listenAddr, sessionID, "snapshot")
	if err != nil {
		return "", 0, 0, err
	}
	defer conn.Close()

	cols, rows := 80, 24
	var out strings.Builder
	for {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				if out.Len() > 0 {
					return out.String(), cols, rows, nil
				}
				continue
			}
			if out.Len() > 0 {
				return out.String(), cols, rows, nil
			}
			return "", cols, rows, normalizeTerminalReadError(err)
		}
		switch msgType {
		case websocket.TextMessage:
			if handled, _, parseErr := parseServerMessage(data); parseErr != nil {
				return "", cols, rows, parseErr
			} else if handled {
				cols, rows = parseAttachRoleDimensions(data, cols, rows)
			}
		case websocket.BinaryMessage:
			out.Write(data)
		}
	}
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