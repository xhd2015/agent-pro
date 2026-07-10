package ttywatch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

const (
	detachByte = 0x1d
)

// TerminalWebSocketURL builds the WebSocket URL for a terminal session.
func TerminalWebSocketURL(listenAddr, sessionID, attachMode string) (string, error) {
	return terminalWebSocketURL(listenAddr, sessionID, attachMode)
}

// DialTerminal connects to a live terminal session and consumes the handshake.
func DialTerminal(listenAddr, sessionID, attachMode string) (*websocket.Conn, error) {
	return dialTerminal(listenAddr, sessionID, attachMode)
}

func dialTerminal(listenAddr, sessionID, attachMode string) (*websocket.Conn, error) {
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
	if err := consumeSessionHandshake(conn, sessionID); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func consumeSessionHandshake(conn *websocket.Conn, knownSessionID string) error {
	deadline := time.Now().Add(10 * time.Second)
	_ = conn.SetReadDeadline(deadline)
	for time.Now().Before(deadline) {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				if knownSessionID != "" {
					_ = conn.SetReadDeadline(time.Time{})
					return nil
				}
				return fmt.Errorf("timeout waiting for session handshake")
			}
			return err
		}
		if msgType == websocket.TextMessage {
			handled, sessionID, err := parseServerMessage(data)
			if err != nil {
				return err
			}
			if handled && sessionID != "" {
				_ = conn.SetReadDeadline(time.Time{})
				return nil
			}
		} else if knownSessionID != "" {
			_ = conn.SetReadDeadline(time.Time{})
			return nil
		}
	}
	if knownSessionID != "" {
		_ = conn.SetReadDeadline(time.Time{})
		return nil
	}
	return fmt.Errorf("timeout waiting for session handshake")
}

var altScreenExitPrefix = []byte("\x1b[?1049l\x1b[0m")

type wsWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsWriter) writeMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(messageType, data)
}

func (w *wsWriter) writeJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return w.writeMessage(websocket.TextMessage, data)
}

func (w *wsWriter) close(code int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	msg := websocket.FormatCloseMessage(code, "")
	return w.conn.WriteControl(websocket.CloseMessage, msg, time.Time{})
}

// attachStdoutWriter writes PTY bytes to interactive terminals with normalizeTTYOutput
// so bare LF advances to column 0 (CRLF); standalone \r in-place redraws are preserved.
// When stdout is not a TTY, it emulates \r overwrite for plain-text capture instead of
// stripping \r (which smears redrawn shell errors). The alternate-screen exit
// prefix is followed by a newline so scrollback text appears on the next line.
type attachStdoutWriter struct {
	w       io.Writer
	rawTTY  bool
	lineBuf []byte
}

func (a *attachStdoutWriter) Write(p []byte) (int, error) {
	if bytes.Equal(p, altScreenExitPrefix) {
		debugLogf("attachStdoutWriter alt-screen-exit-prefix dropped rawTTY=%v", a.rawTTY)
		if _, err := a.w.Write([]byte{'\n'}); err != nil {
			return 0, err
		}
		debugLogBytes("attachStdoutWriter wrote newline after alt-screen prefix", []byte{'\n'})
		return len(p), nil
	}
	if a.rawTTY {
		debugLogBytes("attachStdoutWriter rawTTY in", p)
		out := normalizeTTYOutput(p)
		debugLogBytes("attachStdoutWriter rawTTY out", out)
		if _, err := a.w.Write(out); err != nil {
			return 0, err
		}
		return len(p), nil
	}
	debugLogBytes("attachStdoutWriter pipeCapture in", p)
	n, err := a.writePipeCapture(p)
	if err == nil {
		debugLogf("attachStdoutWriter pipeCapture ok in_len=%d", len(p))
	}
	return n, err
}

func normalizeCRLF(b []byte) []byte {
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	b = bytes.ReplaceAll(b, []byte("\r"), []byte("\n"))
	return b
}

// normalizeTTYOutput prepares bytes for interactive TTY stdout. LF-only newlines
// are expanded to CRLF so the cursor returns to column 0 after each line; without
// CR, a short line like "yes" leaves the cursor at column 3 and the next line
// appears indented. Standalone carriage returns are preserved for in-place redraws.
func normalizeTTYOutput(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	var out bytes.Buffer
	out.Grow(len(b) + bytes.Count(b, []byte{'\n'}))
	for i, c := range b {
		if c != '\n' {
			out.WriteByte(c)
			continue
		}
		if i > 0 && b[i-1] == '\r' {
			out.WriteByte('\n')
			continue
		}
		out.WriteString("\r\n")
	}
	return out.Bytes()
}

func (a *attachStdoutWriter) writePipeCapture(p []byte) (int, error) {
	buf := normalizeCRLF(append(a.lineBuf, p...))
	a.lineBuf = nil

	for {
		idx := bytes.IndexByte(buf, '\n')
		if idx < 0 {
			a.lineBuf = buf
			return len(p), nil
		}
		line := bytes.TrimLeft(buf[:idx], " \t")
		if len(line) > 0 {
			if _, err := a.w.Write(line); err != nil {
				return 0, err
			}
		}
		if _, err := a.w.Write([]byte{'\n'}); err != nil {
			return 0, err
		}
		buf = buf[idx+1:]
	}
}

type detachReader struct {
	r        io.Reader
	detached bool
}

func (d *detachReader) Read(p []byte) (int, error) {
	n, err := d.r.Read(p)
	if n <= 0 {
		return n, err
	}
	if idx := indexByte(p[:n], detachByte); idx >= 0 {
		d.detached = true
		// Return bytes before the detach key (may be zero). Callers must check
		// d.detached when n==0 with EOF — a lone Ctrl-] is a successful detach.
		return idx, io.EOF
	}
	return n, err
}

// envOpenAttachInstant is set by tests (AGENT_RUN_OPEN_ATTACH_INSTANT=1) so
// run --open auto-attach returns without a controlling TTY.
const envOpenAttachInstant = "AGENT_RUN_OPEN_ATTACH_INSTANT"

// AttachWriter attaches stdin/stdout to a live terminal session.
// Returns detached=true when the user pressed Ctrl-].
func AttachWriter(listenAddr, sessionID, attachMode string) (detached bool, err error) {
	// CI / doctest hook for agent-run run --open: skip interactive attach.
	if os.Getenv(envOpenAttachInstant) == "1" {
		debugLogf("attachWriter instant return session=%s listen=%s (AGENT_RUN_OPEN_ATTACH_INSTANT=1)", sessionID, listenAddr)
		return true, nil
	}
	stdoutFile := os.Stdout
	rawTTY := term.IsTerminal(int(stdoutFile.Fd()))
	cols, rows := TTYAttachTerminalSize(stdoutFile)

	sink, cleanup, err := NewTTYAttachSink()
	if err != nil {
		return false, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	debugLogf("attachWriter session=%s listen=%s rawTTY=%v cols=%d rows=%d attach_mode=%s",
		sessionID, listenAddr, rawTTY, cols, rows, attachMode)

	cfg := AttachRelayConfig{
		ExitOnTerminalExit:           true,
		SkipScreenSnapshotConversion: false,
		Cols:                         cols,
		Rows:                         rows,
		OnConnect:                    TTYAttachOnConnect(stdoutFile),
	}
	runErr := AttachRelay(context.Background(), listenAddr, sessionID, attachMode, cfg, sink)
	if sink.Detached() {
		debugLogf("attachWriter detached session=%s", sessionID)
		return true, nil
	}
	debugLogf("attachWriter done session=%s err=%v", sessionID, runErr)
	return false, runErr
}

const terminalExitMarkerText = "\r\n[Terminal exited]\r\n"

func relayTerminalOutput(conn *websocket.Conn, stdout io.Writer, exitOnTerminalExit bool, cols, rows int, observerMode bool, skipScreenSnapshotConversion bool) error {
	screenSnapshotPending := true
	exitMarkerWritten := false
	if flusher, ok := stdout.(observerFlusher); ok {
		defer func() { _ = flusher.Flush() }()
	}
	debugLogf("relayTerminalOutput start exitOnTerminalExit=%v observerMode=%v skipScreenSnapshotConversion=%v cols=%d rows=%d", exitOnTerminalExit, observerMode, skipScreenSnapshotConversion, cols, rows)
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			debugLogf("relayTerminalOutput read err=%v", err)
			return err
		}
		switch msgType {
		case websocket.BinaryMessage:
			debugLogBytes("relayTerminalOutput ws binary in", data)
			converted := false
			prependNL := false
			if observerMode {
				if handler, ok := stdout.(observerBinaryHandler); ok {
					if err := handler.WriteObserverBinary(data); err != nil {
						return err
					}
					continue
				}
				data = RenderObserverFrame(data, cols, rows)
				if len(data) == 0 {
					continue
				}
				converted = true
			} else if !skipScreenSnapshotConversion && screenSnapshotPending && isScreenSnapshotFrame(data) {
				screenSnapshotPending = false
				debugLogf("relayTerminalOutput screen snapshot frame cols=%d rows=%d", cols, rows)
				if text, ok := screenSnapshotToText(data, cols, rows); ok {
					converted = true
					debugLogBytes("relayTerminalOutput screen snapshot text", text)
					prependNL = shouldPrependSnapshotNewline(text)
					if prependNL {
						data = append([]byte{'\n'}, text...)
					} else {
						data = text
					}
					debugLogf("relayTerminalOutput snapshot converted prependNL=%v exitMarkerInText=%v",
						prependNL, bytes.Contains(text, []byte("[Terminal exited]")))
					if bytes.Contains(text, []byte("[Terminal exited]")) {
						exitMarkerWritten = true
					}
				} else {
					debugLogf("relayTerminalOutput screen snapshot toText failed")
				}
			} else if screenSnapshotPending {
				debugLogf("relayTerminalOutput binary not screen snapshot pending=%v isSnapshot=%v",
					screenSnapshotPending, isScreenSnapshotFrame(data))
			}
			debugLogBytes("relayTerminalOutput ws binary out", data)
			if _, err := stdout.Write(data); err != nil {
				return err
			}
			if bytes.Contains(data, []byte("[Terminal exited]")) {
				exitMarkerWritten = true
			}
			if exitOnTerminalExit && isTerminalExitMarker(data) {
				debugLogf("relayTerminalOutput exit on binary converted=%v exitMarkerWritten=%v", converted, exitMarkerWritten)
				return nil
			}
		case websocket.TextMessage:
			debugLogBytes("relayTerminalOutput ws text in", data)
			if isTerminalExitMarker(data) {
				debugLogf("relayTerminalOutput text exit marker exitMarkerWritten=%v", exitMarkerWritten)
				if !exitMarkerWritten {
					debugLogBytes("relayTerminalOutput writing terminalExitMarkerText", []byte(terminalExitMarkerText))
					if _, err := stdout.Write([]byte(terminalExitMarkerText)); err != nil {
						return err
					}
					exitMarkerWritten = true
				}
				if exitOnTerminalExit {
					return nil
				}
				continue
			}
			handled, _, err := parseServerMessage(data)
			if err != nil {
				return err
			}
			debugLogf("relayTerminalOutput text handled=%v", handled)
			if !handled {
				debugLogBytes("relayTerminalOutput ws text out", data)
				if _, err := stdout.Write(data); err != nil {
					return err
				}
			}
		}
	}
}

func forwardInputWithDetach(writer *wsWriter, stdin io.Reader) (detached bool, err error) {
	buf := make([]byte, 4096)
	for {
		n, readErr := stdin.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if idx := indexByte(chunk, detachByte); idx >= 0 {
				if idx > 0 {
					if writeErr := writer.writeMessage(websocket.BinaryMessage, chunk[:idx]); writeErr != nil {
						return false, writeErr
					}
				}
				return true, nil
			}
			if writeErr := writer.writeMessage(websocket.BinaryMessage, chunk); writeErr != nil {
				return false, writeErr
			}
		}
		if readErr != nil {
			return false, readErr
		}
	}
}

func isTerminalExitMarker(data []byte) bool {
	return strings.Contains(string(data), "[Terminal exited]")
}

func sendTerminalResize(writer *wsWriter, stdout *os.File) error {
	cols, rows, err := term.GetSize(int(stdout.Fd()))
	if err != nil {
		return nil
	}
	return writer.writeJSON(map[string]any{"type": "resize", "cols": cols, "rows": rows})
}

func indexByte(b []byte, target byte) int {
	for i, v := range b {
		if v == target {
			return i
		}
	}
	return -1
}