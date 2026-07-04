package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

const (
	detachByte      = 0x1d
	watchDetachByte = 0x03
)

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

type serverMessage struct {
	Type      string `json:"type"`
	Message   string `json:"message,omitempty"`
	SessionID string `json:"session_id,omitempty"`
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

var altScreenExitPrefix = []byte("\x1b[?1049l\x1b[0m")

// Detach cleanup restores the observer terminal after grok-like raw TTY modes
// (alternate screen, kitty keyboard protocol, mouse tracking).
const observerTTYDetachCleanup = "\x1b[?1049l\x1b[<u\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l\x1b[0m"

func writeObserverTTYDetachCleanup(w io.Writer) {
	_, _ = io.WriteString(w, observerTTYDetachCleanup)
}

func isTerminalExitMarker(data []byte) bool {
	return strings.Contains(string(data), "[Terminal exited]")
}

// shouldPrependSnapshotNewline adds a leading newline for live screen snapshots so
// scrollback does not collide with the host prompt, but not for short completed
// output (e.g. echo yes) where a blank first line is user-visible smear.
func shouldPrependSnapshotNewline(text []byte) bool {
	return !bytes.Contains(text, []byte("[Terminal exited]"))
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
		// Scrollback replay prefixes alternate-screen exit before dumping
		// history; on a cleared terminal that paints blank lines. Drop it
		// and emit a single newline so following scrollback stays on a fresh line.
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
		return idx, io.EOF
	}
	return n, err
}

func attachWriter(listenAddr, sessionID, attachMode string) (detached bool, err error) {
	conn, err := dialTerminal(listenAddr, sessionID, attachMode)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	stdoutFile := os.Stdout
	rawTTY := term.IsTerminal(int(stdoutFile.Fd()))

	stdin := &detachReader{r: os.Stdin}
	stdinFile := os.Stdin
	stdinIsTTY := term.IsTerminal(int(stdinFile.Fd()))
	var oldStdinState *term.State
	if stdinIsTTY {
		if state, err := term.MakeRaw(int(stdinFile.Fd())); err == nil {
			oldStdinState = state
			defer term.Restore(int(stdinFile.Fd()), oldStdinState)
		}
	}

	writer := &wsWriter{conn: conn}
	if rawTTY {
		_ = sendTerminalResize(writer, stdoutFile)
		if term.IsTerminal(int(stdinFile.Fd())) {
			sigWinch := make(chan os.Signal, 1)
			signal.Notify(sigWinch, syscall.SIGWINCH)
			defer signal.Stop(sigWinch)
			go func() {
				for range sigWinch {
					_ = sendTerminalResize(writer, stdoutFile)
				}
			}()
		}
	}

	cols, rows := 80, 24
	if rawTTY {
		if c, r, err := term.GetSize(int(stdoutFile.Fd())); err == nil {
			cols, rows = c, r
		}
	}

	debugLogf("attachWriter session=%s listen=%s rawTTY=%v cols=%d rows=%d attach_mode=%s",
		sessionID, listenAddr, rawTTY, cols, rows, attachMode)

	out := &attachStdoutWriter{w: stdoutFile, rawTTY: rawTTY}
	readerErrCh := make(chan error, 1)
	go func() {
		readerErrCh <- relayTerminalOutput(conn, out, true, cols, rows, false, false)
	}()

	var runErr error
	if stdinIsTTY {
		stdinErrCh := make(chan error, 1)
		go func() {
			detached, err := forwardInputWithDetach(writer, stdin)
			if detached {
				stdin.detached = true
				err = nil
			}
			stdinErrCh <- err
		}()

		select {
		case err := <-readerErrCh:
			runErr = normalizeTerminalReadError(err)
		case err := <-stdinErrCh:
			if err != nil && err != io.EOF {
				runErr = err
			}
		}
	} else {
		runErr = normalizeTerminalReadError(<-readerErrCh)
	}

	_ = writer.close(websocket.CloseNormalClosure)
	if stdin.detached {
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
				data = renderObserverFrame(data, cols, rows)
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

const kittyDetachMaxPending = 32

// kittyCtrlCCodepoint is ETX (0x03), the unicode key code many terminals report for
// Ctrl-C under the kitty keyboard protocol.
const kittyCtrlCCodepoint = 3

// kittyCtrlCLetterCodepoint is lowercase 'c'; iTerm2 and others send Ctrl-C as key 99
// with the ctrl modifier instead of key 3 (ETX).
const kittyCtrlCLetterCodepoint = 99

// kittyCtrlModifierBit is the ctrl modifier bit in kitty keyboard protocol modifier fields.
const kittyCtrlModifierBit = 4

func drainObserverInput(stdin io.Reader) (detached bool, err error) {
	buf := make([]byte, 4096)
	var pending []byte
	for {
		n, readErr := stdin.Read(buf)
		if n > 0 {
			debugLogBytes("drainObserverInput read", buf[:n])
			pending = append(pending, buf[:n]...)
			if observerInputDetach(pending) {
				debugLogf("drainObserverInput detach pending_len=%d", len(pending))
				return true, nil
			}
			pending = trimObserverInputPending(pending)
		}
		if readErr != nil {
			return false, readErr
		}
	}
}

func observerInputDetach(b []byte) bool {
	if indexByte(b, watchDetachByte) >= 0 {
		return true
	}
	return containsKittyCtrlC(b)
}

func trimObserverInputPending(pending []byte) []byte {
	if len(pending) > kittyDetachMaxPending {
		pending = pending[len(pending)-kittyDetachMaxPending:]
	}
	if idx := bytes.LastIndexByte(pending, 0x1b); idx >= 0 {
		if idx > 0 && len(pending)-idx < kittyDetachMaxPending {
			pending = pending[idx:]
		}
	}
	return pending
}

// containsKittyCtrlC reports whether b contains a complete kitty keyboard protocol
// Ctrl-C sequence (CSI unicode-key-code ; modifiers u with key 3 or 99 and ctrl held).
func containsKittyCtrlC(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if b[i] != 0x1b || i+1 >= len(b) || b[i+1] != '[' {
			continue
		}
		j := i + 2
		for j < len(b) && b[j] >= 0x30 && b[j] <= 0x3f {
			j++
		}
		if j >= len(b) || b[j] != 'u' {
			continue
		}
		if kittyCSIParamsAreCtrlC(string(b[i+2 : j])) {
			return true
		}
	}
	return false
}

func kittyCSIParamsAreCtrlC(params string) bool {
	semi := strings.LastIndexByte(params, ';')
	if semi < 0 {
		return false
	}
	keyPart := params[:semi]
	modPart := params[semi+1:]
	if colon := strings.IndexByte(modPart, ':'); colon >= 0 {
		modPart = modPart[:colon]
	}
	keyPart = strings.SplitN(keyPart, ":", 2)[0]
	keyCode, err := strconv.Atoi(keyPart)
	if err != nil || !kittyCSIKeyIsCtrlC(keyCode) {
		return false
	}
	mods, err := strconv.Atoi(modPart)
	if err != nil {
		return false
	}
	return mods&kittyCtrlModifierBit != 0
}

func kittyCSIKeyIsCtrlC(keyCode int) bool {
	return keyCode == kittyCtrlCCodepoint || keyCode == kittyCtrlCLetterCodepoint
}

func indexByte(b []byte, target byte) int {
	for i, v := range b {
		if v == target {
			return i
		}
	}
	return -1
}

func sendTerminalResize(writer *wsWriter, stdout *os.File) error {
	cols, rows, err := term.GetSize(int(stdout.Fd()))
	if err != nil {
		return nil
	}
	return writer.writeJSON(map[string]any{"type": "resize", "cols": cols, "rows": rows})
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

type attachRoleMessage struct {
	Type       string `json:"type"`
	AttachRole string `json:"attach_role"`
	Cols       int    `json:"cols"`
	Rows       int    `json:"rows"`
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

func streamObserver(listenAddr, sessionID string, stdout io.Writer) error {
	conn, err := dialTerminal(listenAddr, sessionID, "observer")
	if err != nil {
		return err
	}
	defer conn.Close()

	cols, rows := 80, 24
	stdoutFile, hasTTY := stdout.(*os.File)
	rawTTY := hasTTY && term.IsTerminal(int(stdoutFile.Fd()))
	if rawTTY {
		if c, r, err := term.GetSize(int(stdoutFile.Fd())); err == nil {
			cols, rows = c, r
		}
		writer := &wsWriter{conn: conn}
		_ = sendTerminalResize(writer, stdoutFile)
		sigWinch := make(chan os.Signal, 1)
		signal.Notify(sigWinch, syscall.SIGWINCH)
		defer signal.Stop(sigWinch)
		go func() {
			for range sigWinch {
				_ = sendTerminalResize(writer, stdoutFile)
			}
		}()
	}

	var out io.Writer = stdout
	observerMode := false
	skipScreenSnapshotConversion := false
	if rawTTY {
		out = &attachStdoutWriter{w: stdoutFile, rawTTY: true}
		skipScreenSnapshotConversion = true
	} else {
		out = newObserverPipeWriter(stdout, cols, rows)
		observerMode = true
	}

	if !rawTTY {
		return relayTerminalOutput(conn, out, false, cols, rows, observerMode, skipScreenSnapshotConversion)
	}

	stdinFile := os.Stdin
	var oldStdinState *term.State
	stdinRaw := false
	if term.IsTerminal(int(stdinFile.Fd())) {
		if state, err := term.MakeRaw(int(stdinFile.Fd())); err == nil {
			oldStdinState = state
			stdinRaw = true
		}
	}
	restoreStdinBeforeObserverCleanup := func() {
		if oldStdinState == nil {
			return
		}
		_ = term.Restore(int(stdinFile.Fd()), oldStdinState)
		oldStdinState = nil
	}
	defer restoreStdinBeforeObserverCleanup()

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, syscall.SIGINT)
	defer signal.Stop(sigintCh)
	debugLogf("streamObserver session=%s stdinRaw=%v", sessionID, stdinRaw)

	var (
		detached    atomic.Bool
		cleanupOnce sync.Once
	)
	detachCleanup := func() {
		if !rawTTY {
			return
		}
		cleanupOnce.Do(func() {
			restoreStdinBeforeObserverCleanup()
			writeObserverTTYDetachCleanup(stdoutFile)
		})
	}

	readerErrCh := make(chan error, 1)
	go func() {
		readerErrCh <- relayTerminalOutput(conn, out, false, cols, rows, observerMode, skipScreenSnapshotConversion)
	}()

	stdinErrCh := make(chan error, 1)
	go func() {
		d, err := drainObserverInput(stdinFile)
		if d {
			detached.Store(true)
			detachCleanup()
			err = nil
		}
		stdinErrCh <- err
	}()

	for {
		select {
		case err := <-readerErrCh:
			if detached.Load() {
				return nil
			}
			return normalizeTerminalReadError(err)
		case err := <-stdinErrCh:
			if detached.Load() {
				return nil
			}
			if err != nil && err != io.EOF {
				return err
			}
		case <-sigintCh:
			debugLogf("streamObserver detach via SIGINT session=%s", sessionID)
			detached.Store(true)
			detachCleanup()
			return nil
		}
	}
}
