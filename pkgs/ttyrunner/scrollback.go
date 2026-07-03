package ttyrunner

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// FetchScrollbackSnapshot reads scrollback via attach_mode=snapshot WebSocket probe.
func FetchScrollbackSnapshot(listenAddr, sessionID string) []byte {
	return fetchScrollbackSnapshot(listenAddr, sessionID)
}

func fetchScrollbackSnapshot(listenAddr, sessionID string) []byte {
	if listenAddr == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := url.URL{
		Scheme: "ws",
		Host:   listenAddr,
		Path:   "/api/terminal",
	}
	q := wsURL.Query()
	q.Set("session_id", sessionID)
	q.Set("attach_mode", "snapshot")
	wsURL.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL.String(), nil)
	if err != nil {
		return nil
	}
	defer conn.Close()

	deadline := time.Now().Add(2 * time.Second)
	var out strings.Builder
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		out.Write(msg)
	}
	return []byte(out.String())
}

// WaitUntilWritable polls CheckWritable until ready or timeout.
func WaitUntilWritable(provider Provider, listenAddr, sessionID string, timeout time.Duration) WritableStatus {
	deadline := time.Now().Add(timeout)
	var last WritableStatus
	for time.Now().Before(deadline) {
		scrollback := fetchScrollbackSnapshot(listenAddr, sessionID)
		if provider.CheckWritable != nil {
			last = provider.CheckWritable(scrollback)
		}
		if last.Ready {
			return last
		}
		time.Sleep(150 * time.Millisecond)
	}
	if last.Reason == "" {
		last.Reason = "timed out waiting for writable prompt"
	}
	return last
}