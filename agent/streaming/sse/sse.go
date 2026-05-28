package sse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

type Writer struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewWriter(w http.ResponseWriter) *Writer {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	return &Writer{w: w, flusher: flusher}
}

func (s *Writer) Send(data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(s.w, "data: %s\n\n", jsonData)
	s.flusher.Flush()
}

func (s *Writer) WriteEvent(event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(s.w, "event: %s\n", event)
	fmt.Fprintf(s.w, "data: %s\n\n", jsonData)
	s.flusher.Flush()
}

func (s *Writer) SendLog(message string) {
	s.Send(map[string]string{"type": "log", "message": message})
}

func (s *Writer) SendError(message string) {
	s.Send(map[string]string{"type": "error", "message": message})
}

func (s *Writer) SendDone(extra map[string]string) {
	data := map[string]string{"type": "done"}
	for k, v := range extra {
		data[k] = v
	}
	s.Send(data)
}

func (s *Writer) SendStatus(status string, extra map[string]string) {
	data := map[string]string{"type": "status", "status": status}
	for k, v := range extra {
		data[k] = v
	}
	s.Send(data)
}

func (s *Writer) StreamCmd(cmd *exec.Cmd) error {
	return s.StreamCmdFunc(cmd, nil)
}

func (s *Writer) StreamCmdFunc(cmd *exec.Cmd, onLine func(line string) bool) error {
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return fmt.Errorf("failed to start: %v", err)
	}

	waitErr := make(chan error, 1)
	go func() {
		waitErr <- cmd.Wait()
		pw.Close()
	}()

	scanner := bufio.NewScanner(pr)
	scanner.Split(splitLines)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if onLine == nil || onLine(line) {
			s.SendLog(line)
		}
	}
	pr.Close()

	return <-waitErr
}

func splitLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, b := range data {
		if b == '\n' || b == '\r' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
