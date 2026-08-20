package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
	"golang.org/x/term"
)

// errDetached is returned when the user interrupts an interactive attach.
// The remote session is left running.
var errDetached = errors.New("detached")

const (
	attachKeepAliveInterval = 60 * time.Second
	ctrlCByte               = 0x03
)

type attachControlMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

type attachServerMessage struct {
	Type      string `json:"type"`
	Message   string `json:"message,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

type attachWSWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *attachWSWriter) writeMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(messageType, data)
}

func (w *attachWSWriter) writeJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return w.writeMessage(websocket.TextMessage, data)
}

func (w *attachWSWriter) close(code int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	msg := websocket.FormatCloseMessage(code, "")
	return w.conn.WriteControl(websocket.CloseMessage, msg, time.Time{})
}

func (w *attachWSWriter) writePing() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	deadline := time.Now().Add(time.Second)
	return w.conn.WriteControl(websocket.PingMessage, nil, deadline)
}

// attachSession bridges the local TTY to the daemon session.
//
// attach_mode=attach claims roleAttacher so a bare disconnect does not
// stopChild(). SIGINT / Ctrl-C also send detach_keep and return errDetached
// so the remote command keeps running.
func attachSession(c *ptyclient.Client, sessionID string) error {
	wsURL, err := attachWebSocketURL(c.BaseURL, sessionID)
	if err != nil {
		return err
	}

	header := http.Header{}
	if token := strings.TrimSpace(c.AuthToken); token != "" {
		header.Set("Authorization", "Bearer "+token)
	}
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return attachDialError(err, resp)
	}
	defer conn.Close()

	if err := consumeAttachHandshake(conn, sessionID); err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)
	defer signal.Stop(sigCh)

	writer := &attachWSWriter{conn: conn}
	stopKeepAlive := make(chan struct{})
	defer close(stopKeepAlive)
	startAttachKeepAlive(writer, stopKeepAlive)

	stdinFile := os.Stdin
	var oldState *term.State
	if term.IsTerminal(int(stdinFile.Fd())) {
		if state, err := term.MakeRaw(int(stdinFile.Fd())); err == nil {
			oldState = state
			defer term.Restore(int(stdinFile.Fd()), oldState)
		}
	}

	if term.IsTerminal(int(os.Stdout.Fd())) {
		_ = sendAttachResize(writer, os.Stdout)
		sigWinch := make(chan os.Signal, 1)
		signal.Notify(sigWinch, syscall.SIGWINCH)
		defer signal.Stop(sigWinch)
		go func() {
			for range sigWinch {
				_ = sendAttachResize(writer, os.Stdout)
			}
		}()
	}

	readerErrCh := make(chan error, 1)
	go func() {
		readerErrCh <- readAttachOutput(conn, os.Stdout)
	}()

	stdinErrCh := make(chan error, 1)
	go func() {
		stdinErrCh <- forwardAttachInput(writer, os.Stdin)
	}()

	// Poll session list: if the remote child already exited but the WS never
	// delivered "[Terminal exited]", stdin forward would block forever.
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-sigCh:
			return detachAndClose(writer)
		case err := <-readerErrCh:
			return normalizeAttachReadError(err)
		case err := <-stdinErrCh:
			if errors.Is(err, errDetached) {
				return detachAndClose(writer)
			}
			if err != nil && err != io.EOF {
				return err
			}
			_ = writer.close(websocket.CloseNormalClosure)
			return nil
		case <-ticker.C:
			if sessionExited(c, sessionID) {
				_ = writer.close(websocket.CloseNormalClosure)
				return nil
			}
		}
	}
}

func detachAndClose(writer *attachWSWriter) error {
	_ = writer.writeJSON(attachControlMessage{Type: "detach_keep"})
	_ = writer.close(websocket.CloseNormalClosure)
	return errDetached
}

func attachWebSocketURL(base, sessionID string) (string, error) {
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return "", fmt.Errorf("invalid server url %q: %w", base, err)
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported server scheme %q", u.Scheme)
	}
	u.Path = "/api/terminal"
	q := u.Query()
	if strings.TrimSpace(sessionID) != "" {
		q.Set("session_id", sessionID)
	}
	// roleAttacher: input/resize allowed; disconnect does not reap the child.
	q.Set("attach_mode", "attach")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func consumeAttachHandshake(conn *websocket.Conn, knownSessionID string) error {
	deadline := time.Now().Add(10 * time.Second)
	_ = conn.SetReadDeadline(deadline)
	defer conn.SetReadDeadline(time.Time{})
	for time.Now().Before(deadline) {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				if knownSessionID != "" {
					return nil
				}
				return fmt.Errorf("timeout waiting for session_id message")
			}
			return err
		}
		if msgType == websocket.TextMessage {
			handled, sessionID, err := parseAttachServerMessage(data)
			if err != nil {
				return err
			}
			if handled && sessionID != "" {
				return nil
			}
		} else if knownSessionID != "" {
			return nil
		}
	}
	if knownSessionID != "" {
		return nil
	}
	return fmt.Errorf("timeout waiting for session_id message")
}

func parseAttachServerMessage(data []byte) (handled bool, sessionID string, err error) {
	var msg attachServerMessage
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

func startAttachKeepAlive(writer *attachWSWriter, stop <-chan struct{}) {
	go func() {
		t := time.NewTicker(attachKeepAliveInterval)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				if err := writer.writePing(); err != nil {
					return
				}
			}
		}
	}()
}

func sendAttachResize(writer *attachWSWriter, stdout *os.File) error {
	cols, rows, err := term.GetSize(int(stdout.Fd()))
	if err != nil {
		return nil
	}
	return writer.writeJSON(attachControlMessage{Type: "resize", Cols: cols, Rows: rows})
}

func forwardAttachInput(writer *attachWSWriter, stdin io.Reader) error {
	buf := make([]byte, 4096)
	for {
		n, err := stdin.Read(buf)
		if n > 0 {
			if idx := indexByte(buf[:n], ctrlCByte); idx >= 0 {
				if idx > 0 {
					if writeErr := writer.writeMessage(websocket.BinaryMessage, buf[:idx]); writeErr != nil {
						return writeErr
					}
				}
				return errDetached
			}
			if writeErr := writer.writeMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			return err
		}
	}
}

func readAttachOutput(conn *websocket.Conn, stdout io.Writer) error {
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		switch msgType {
		case websocket.BinaryMessage:
			// ptywrap may replay its terminal-exit lifecycle trailer inside the
			// binary scrollback frame. It is transport metadata, not command
			// output, so avoid printing it before `run` emits the session ID.
			// The server may split the trailer across several binary frames.
			data = bytes.ReplaceAll(data, []byte("\x1b[?1049l"), nil)
			data = bytes.ReplaceAll(data, []byte("\x1b[0m"), nil)
			data = bytes.ReplaceAll(data, []byte("[Terminal exited]"), nil)
			if _, err := stdout.Write(data); err != nil {
				return err
			}
		case websocket.TextMessage:
			if strings.Contains(string(data), "[Terminal exited]") {
				return nil
			}
			handled, _, err := parseAttachServerMessage(data)
			if err != nil {
				return err
			}
			if !handled {
				if _, err := stdout.Write(data); err != nil {
					return err
				}
			}
		}
	}
}

func normalizeAttachReadError(err error) error {
	if err == nil {
		return nil
	}
	if ce, ok := err.(*websocket.CloseError); ok {
		switch ce.Code {
		case websocket.CloseNormalClosure, websocket.CloseGoingAway, 4000:
			return nil
		}
		return fmt.Errorf("terminal closed: %s", ce.Text)
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		return nil
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		return nil
	}
	return err
}

func attachDialError(err error, resp *http.Response) error {
	if resp == nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		return fmt.Errorf("terminal connect failed: %s", resp.Status)
	}
	return fmt.Errorf("terminal connect failed: %s: %s", resp.Status, snippet)
}

func indexByte(b []byte, target byte) int {
	for i, v := range b {
		if v == target {
			return i
		}
	}
	return -1
}
