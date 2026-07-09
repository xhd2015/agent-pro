package ttywatch

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

// AttachSink receives PTY output and forwards interactive input to the upstream terminal.
type AttachSink interface {
	OutputWriter() io.Writer
	RunInput(ctx context.Context, upstream *wsWriter) error
	// DrainOutputAfterInput reports whether upstream output should be read briefly
	// after stdin reaches EOF (piped attach clients).
	DrainOutputAfterInput() bool
}

// AttachRelayConfig controls relayTerminalOutput behaviour for an attach session.
type AttachRelayConfig struct {
	ExitOnTerminalExit           bool
	SkipScreenSnapshotConversion bool
	Cols, Rows                   int
	OnConnect                    func(upstream *wsWriter) error
}

// AttachRelay connects to a live terminal and relays output/input through sink.
func AttachRelay(ctx context.Context, listenAddr, sessionID, attachMode string, cfg AttachRelayConfig, sink AttachSink) error {
	conn, err := dialTerminal(listenAddr, sessionID, attachMode)
	if err != nil {
		return err
	}
	defer conn.Close()

	writer := &wsWriter{conn: conn}

	cols, rows := cfg.Cols, cfg.Rows
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	inputCtx, cancelInput := context.WithCancel(ctx)
	defer cancelInput()

	inputErrCh := make(chan error, 1)
	go func() {
		inputErrCh <- sink.RunInput(inputCtx, writer)
	}()

	if cfg.OnConnect != nil {
		if err := cfg.OnConnect(writer); err != nil {
			cancelInput()
			return err
		}
	}

	outErrCh := make(chan error, 1)
	go func() {
		outErrCh <- relayTerminalOutput(
			conn,
			sink.OutputWriter(),
			cfg.ExitOnTerminalExit,
			cols,
			rows,
			false,
			cfg.SkipScreenSnapshotConversion,
		)
	}()

	var runErr error
	select {
	case outErr := <-outErrCh:
		cancelInput()
		runErr = normalizeTerminalReadError(outErr)
	case inErr := <-inputErrCh:
		if inErr != nil && inErr != io.EOF && inErr != context.Canceled {
			runErr = inErr
		}
		if sink.DrainOutputAfterInput() {
			select {
			case outErr := <-outErrCh:
				if runErr == nil {
					runErr = normalizeTerminalReadError(outErr)
				}
			case <-time.After(2 * time.Second):
				_ = writer.close(websocket.CloseNormalClosure)
				select {
				case outErr := <-outErrCh:
					if runErr == nil {
						runErr = normalizeTerminalReadError(outErr)
					}
				case <-time.After(500 * time.Millisecond):
				}
			}
		} else {
			_ = writer.close(websocket.CloseNormalClosure)
			select {
			case outErr := <-outErrCh:
				if runErr == nil {
					runErr = normalizeTerminalReadError(outErr)
				}
			case <-time.After(250 * time.Millisecond):
			}
		}
	}
	return runErr
}

// TTYAttachSink relays attach IO through the process stdin/stdout.
type TTYAttachSink struct {
	stdin      *detachReader
	stdout     *attachStdoutWriter
	stdinIsTTY bool
	detached   bool
}

// NewTTYAttachSink prepares stdin/stdout for interactive attach.
func NewTTYAttachSink() (*TTYAttachSink, func(), error) {
	stdoutFile := os.Stdout
	rawTTY := term.IsTerminal(int(stdoutFile.Fd()))
	stdin := &detachReader{r: os.Stdin}
	stdinFile := os.Stdin
	stdinIsTTY := term.IsTerminal(int(stdinFile.Fd()))
	var cleanup func()
	if stdinIsTTY {
		if state, err := term.MakeRaw(int(stdinFile.Fd())); err == nil {
			cleanup = func() { term.Restore(int(stdinFile.Fd()), state) }
		}
	}
	return &TTYAttachSink{
		stdin:      stdin,
		stdout:     &attachStdoutWriter{w: stdoutFile, rawTTY: rawTTY},
		stdinIsTTY: stdinIsTTY,
	}, cleanup, nil
}

func (s *TTYAttachSink) OutputWriter() io.Writer { return s.stdout }

func (s *TTYAttachSink) DrainOutputAfterInput() bool { return !s.stdinIsTTY }

func (s *TTYAttachSink) RunInput(ctx context.Context, upstream *wsWriter) error {
	_ = ctx
	if s.stdinIsTTY {
		// forwardInputWithDetach and detachReader both detect Ctrl-]. When the
		// detach byte is the first/only input (common for harness "send 0x1d"),
		// detachReader returns (0, EOF) with detached=true and the n>0 branch
		// in forwardInputWithDetach never runs — so we must honor stdin.detached.
		detached, err := forwardInputWithDetach(upstream, s.stdin)
		if detached || s.stdin.detached {
			s.detached = true
			return nil
		}
		return err
	}
	return forwardInput(upstream, s.stdin)
}

func forwardInput(writer *wsWriter, stdin io.Reader) error {
	buf := make([]byte, 4096)
	for {
		n, readErr := stdin.Read(buf)
		if n > 0 {
			if writeErr := writer.writeMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return readErr
		}
	}
}

func (s *TTYAttachSink) Detached() bool { return s.detached }

// WebAttachOnConnect primes browser output with snapshot scrollback. Resize is sent by
// the browser client after the websocket opens so upstream resize acks are not relayed
// before interactive input during attach setup.
func WebAttachOnConnect(listenAddr, sessionID string, out io.Writer, cols, rows int) func(upstream *wsWriter) error {
	return func(upstream *wsWriter) error {
		if cols <= 0 {
			cols = 80
		}
		if rows <= 0 {
			rows = 24
		}
		if text, err := primeSnapshotText(listenAddr, sessionID); err == nil && strings.TrimSpace(text) != "" {
			_, _ = out.Write([]byte(text))
			return nil
		}
		scrollback, snapCols, snapRows, err := readSnapshotScrollback(listenAddr, sessionID)
		if err != nil || scrollback == "" {
			return nil
		}
		if snapCols > 0 {
			cols = snapCols
		}
		if snapRows > 0 {
			rows = snapRows
		}
		text := RenderSnapshotOutput("", scrollback, cols, rows)
		if text == "" {
			text = scrollback
		}
		if text != "" {
			_, _ = out.Write([]byte(text))
		}
		return nil
	}
}

// TTYAttachOnConnect sends an initial resize and watches SIGWINCH when stdout is a TTY.
func TTYAttachOnConnect(stdout *os.File) func(upstream *wsWriter) error {
	return func(upstream *wsWriter) error {
		if !term.IsTerminal(int(stdout.Fd())) {
			return nil
		}
		if err := sendTerminalResize(upstream, stdout); err != nil {
			return err
		}
		sigWinch := make(chan os.Signal, 1)
		signal.Notify(sigWinch, syscall.SIGWINCH)
		go func() {
			for range sigWinch {
				_ = sendTerminalResize(upstream, stdout)
			}
		}()
		return nil
	}
}

// TTYAttachTerminalSize returns the current terminal dimensions when stdout is a TTY.
func TTYAttachTerminalSize(stdout *os.File) (cols, rows int) {
	cols, rows = 80, 24
	if term.IsTerminal(int(stdout.Fd())) {
		if c, r, err := term.GetSize(int(stdout.Fd())); err == nil {
			cols, rows = c, r
		}
	}
	return cols, rows
}

// WebSocketAttachSink relays attach IO through a browser WebSocket connection.
type WebSocketAttachSink struct {
	conn *websocket.Conn
	out  *wsBrowserOutput
}

// NewWebSocketAttachSink creates a sink for browser terminal websocket clients.
func NewWebSocketAttachSink(conn *websocket.Conn) *WebSocketAttachSink {
	return &WebSocketAttachSink{
		conn: conn,
		out:  &wsBrowserOutput{conn: conn},
	}
}

func (s *WebSocketAttachSink) OutputWriter() io.Writer { return s.out }

func (s *WebSocketAttachSink) DrainOutputAfterInput() bool { return false }

func (s *WebSocketAttachSink) RunInput(ctx context.Context, upstream *wsWriter) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		mt, msg, err := s.conn.ReadMessage()
		if err != nil {
			return err
		}
		switch mt {
		case websocket.BinaryMessage:
			if err := upstream.writeMessage(websocket.BinaryMessage, msg); err != nil {
				return err
			}
		case websocket.TextMessage:
			if err := upstream.writeMessage(websocket.TextMessage, msg); err != nil {
				return err
			}
		}
	}
}

type wsBrowserOutput struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsBrowserOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.conn.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// ParseResizeMessage extracts cols/rows from a browser resize JSON message.
func ParseResizeMessage(data []byte) (cols, rows int, ok bool) {
	var msg struct {
		Type string `json:"type"`
		Cols int    `json:"cols"`
		Rows int    `json:"rows"`
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return 0, 0, false
	}
	if msg.Type != "resize" || msg.Cols <= 0 || msg.Rows <= 0 {
		return 0, 0, false
	}
	return msg.Cols, msg.Rows, true
}