package main

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/hinshun/vt10x"
)

const observerScreenFlushDelay = 200 * time.Millisecond

// observerBinaryHandler receives raw PTY binary frames for pipe observer capture.
type observerBinaryHandler interface {
	WriteObserverBinary(data []byte) error
}

// observerFlusher emits the latest accumulated alternate-screen state on close.
type observerFlusher interface {
	Flush() error
}

// observerPipeWriter renders the latest virtual screen for non-TTY watch capture.
// Alternate-screen redraws update internal vt state; the latest screen is emitted
// once after updates settle so pipe capture does not stack duplicate snapshots.
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
	text, ok := renderVTStateToText(o.vt, o.cols, o.rows)
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