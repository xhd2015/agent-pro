package runner

import (
	"io"

	"github.com/xhd2015/agent-pro/agent/agentui/textutil"
)

type Logger struct {
	ch   chan<- string
	buf  []byte
	file io.Writer
}

func NewLogger(ch chan<- string, file io.Writer) *Logger {
	return &Logger{ch: ch, file: file}
}

func (l *Logger) Log(msg string) {
	if msg == "" {
		return
	}
	l.buf = append(l.buf, []byte(msg)...)
	for {
		idx := textutil.IndexByte(l.buf, '\n')
		if idx < 0 {
			break
		}
		line := string(l.buf[:idx])
		l.buf = l.buf[idx+1:]
		if line == "" {
			continue
		}
		if l.file != nil {
			l.file.Write([]byte(line + "\n"))
		}
		formatted := FormatLogLine(line)
		if formatted == "" {
			continue
		}
		select {
		case l.ch <- formatted:
		default:
		}
	}
}
