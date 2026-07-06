package ttywatch

import (
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSAttachClient is a lightweight WebSocket attach helper for doctest harnesses.
type WSAttachClient struct {
	Conn     *websocket.Conn
	output   strings.Builder
	outputMu sync.Mutex
	Role     string
}

// DialPTYAttach connects to a ptywrap session with the given attach mode.
func DialPTYAttach(listenAddr, sessionID, attachMode string) (*WSAttachClient, error) {
	u, err := url.Parse("http://" + listenAddr)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/api/terminal"
	q := u.Query()
	q.Set("session_id", sessionID)
	if attachMode != "" {
		q.Set("attach_mode", attachMode)
	}
	u.RawQuery = q.Encode()
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	client := &WSAttachClient{Conn: conn, Role: attachMode}
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			client.outputMu.Lock()
			client.output.Write(msg)
			client.outputMu.Unlock()
		}
	}()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, _ = conn.ReadMessage()
	_ = conn.SetReadDeadline(time.Time{})
	return client, nil
}

// Output returns accumulated PTY output from the attach client.
func (c *WSAttachClient) Output() string {
	c.outputMu.Lock()
	defer c.outputMu.Unlock()
	return c.output.String()
}

// TryWriteInput sends binary input when the attach role allows writing.
func (c *WSAttachClient) TryWriteInput(data []byte) error {
	return c.Conn.WriteMessage(websocket.BinaryMessage, data)
}

// TryResize sends a resize control message.
func (c *WSAttachClient) TryResize(cols, rows int) error {
	msg, _ := json.Marshal(map[string]any{"type": "resize", "cols": cols, "rows": rows})
	return c.Conn.WriteMessage(websocket.TextMessage, msg)
}

// Close closes the WebSocket connection.
func (c *WSAttachClient) Close() {
	_ = c.Conn.Close()
}