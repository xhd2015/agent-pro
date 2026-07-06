package ttywatch

import (
	"bytes"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hinshun/vt10x"
	"golang.org/x/term"
)

const (
	watchDetachByte          = 0x03
	kittyDetachMaxPending    = 32
	kittyCtrlCCodepoint      = 3
	kittyCtrlCLetterCodepoint = 99
	kittyCtrlModifierBit     = 4
	observerScreenFlushDelay = 200 * time.Millisecond
)

// Detach cleanup restores the observer terminal after grok-like raw TTY modes
// (alternate screen, kitty keyboard protocol, mouse tracking).
const observerTTYDetachCleanup = "\x1b[?1049l\x1b[<u\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l\x1b[0m"

func writeObserverTTYDetachCleanup(w io.Writer) {
	_, _ = io.WriteString(w, observerTTYDetachCleanup)
}

// observerBinaryHandler receives raw PTY binary frames for pipe observer capture.
type observerBinaryHandler interface {
	WriteObserverBinary(data []byte) error
}

// observerFlusher emits the latest accumulated alternate-screen state on close.
type observerFlusher interface {
	Flush() error
}

// observerPipeWriter renders the latest virtual screen for non-TTY watch capture.
type observerPipeWriter struct {
	dest        io.Writer
	cols, rows  int
	vt          vt10x.Terminal
	altScreen   bool
	screenDirty bool
	lineBuf     []byte
	mu          sync.Mutex
	flushTimer  *time.Timer
}

func newObserverPipeWriter(w io.Writer, cols, rows int) *observerPipeWriter {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	return &observerPipeWriter{
		dest: w,
		cols: cols,
		rows: rows,
		vt:   vt10x.New(vt10x.WithSize(cols, rows)),
	}
}

func (o *observerPipeWriter) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.altScreen {
		return len(p), nil
	}
	if err := o.writePlainToDestLocked(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (o *observerPipeWriter) WriteObserverBinary(data []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if bytes.Contains(data, []byte("\x1b[?1049h")) {
		o.altScreen = true
	}
	if bytes.Contains(data, []byte("\x1b[?1049l")) {
		o.stopFlushTimerLocked()
		if o.altScreen && o.screenDirty {
			if err := o.flushScreenLocked(); err != nil {
				return err
			}
		}
		o.altScreen = false
	}

	if _, err := o.vt.Write(data); err != nil {
		return err
	}

	if o.altScreen {
		o.screenDirty = true
		o.scheduleFlushLocked()
		return nil
	}
	return o.writePlainToDestLocked(data)
}

func (o *observerPipeWriter) Flush() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.stopFlushTimerLocked()
	return o.flushScreenLocked()
}

func (o *observerPipeWriter) scheduleFlushLocked() {
	if o.flushTimer != nil {
		o.flushTimer.Stop()
	}
	o.flushTimer = time.AfterFunc(observerScreenFlushDelay, func() {
		o.mu.Lock()
		defer o.mu.Unlock()
		_ = o.flushScreenLocked()
	})
}

func (o *observerPipeWriter) stopFlushTimerLocked() {
	if o.flushTimer != nil {
		o.flushTimer.Stop()
		o.flushTimer = nil
	}
}

func (o *observerPipeWriter) flushScreenLocked() error {
	if !o.screenDirty {
		return nil
	}
	text, ok := RenderVTStateToText(o.vt, o.cols, o.rows)
	o.screenDirty = false
	if !ok {
		return nil
	}
	cleaned := SanitizeForPrint(string(text))
	if cleaned == "" {
		return nil
	}
	_, err := io.WriteString(o.dest, cleaned)
	return err
}

func (o *observerPipeWriter) writePlainToDestLocked(p []byte) error {
	buf := normalizeCRLF(append(o.lineBuf, p...))
	o.lineBuf = nil

	for {
		idx := bytes.IndexByte(buf, '\n')
		if idx < 0 {
			o.lineBuf = buf
			return nil
		}
		line := bytes.TrimLeft(buf[:idx], " \t")
		if len(line) > 0 {
			cleaned := SanitizeForPrint(string(line))
			if cleaned != "" {
				if _, err := io.WriteString(o.dest, cleaned); err != nil {
					return err
				}
			}
		}
		if _, err := io.WriteString(o.dest, "\n"); err != nil {
			return err
		}
		buf = buf[idx+1:]
	}
}

// StreamObserver streams readonly terminal output to w.
func StreamObserver(listenAddr, sessionID string, w io.Writer) error {
	conn, err := dialTerminal(listenAddr, sessionID, "observer")
	if err != nil {
		return err
	}
	defer conn.Close()

	cols, rows := 80, 24
	stdoutFile, hasTTY := w.(*os.File)
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

	var out io.Writer = w
	observerMode := false
	skipScreenSnapshotConversion := false
	if rawTTY {
		out = &attachStdoutWriter{w: stdoutFile, rawTTY: true}
		skipScreenSnapshotConversion = true
	} else {
		out = newObserverPipeWriter(w, cols, rows)
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