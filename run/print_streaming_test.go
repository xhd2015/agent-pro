package run

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/trace"
)

func TestWriteHumanTraceFollowsRunningSession(t *testing.T) {
	dir := t.TempDir()

	metaPath := filepath.Join(dir, "metadata.json")
	logPath := filepath.Join(dir, "events.jsonl")

	initialMeta := trace.AgentTraceMetadata{
		ID:        "test-session",
		Status:    "running",
		Command:   "test-cmd",
		LogPath:   logPath,
		CreatedAt: time.Now().Format(time.RFC3339Nano),
		UpdatedAt: time.Now().Format(time.RFC3339Nano),
	}
	writeJSON(t, metaPath, initialMeta)

	initialEvents := []string{
		`{"type":"item.completed","item":{"id":"msg_1","type":"agent_message","text":"hello"}}`,
	}
	writeLines(t, logPath, initialEvents)

	source := trace.NewSessionDirSource(dir)
	summaries, err := source.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	summary := summaries[0]
	detail, err := source.Get(summary.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	pr, pw := io.Pipe()
	done := make(chan error, 1)

	outputCh := make(chan string, 100)
	readDone := make(chan struct{})
	var allOutput bytes.Buffer
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				allOutput.Write(buf[:n])
				outputCh <- chunk
			}
			if err != nil {
				break
			}
		}
		close(readDone)
	}()

	go func() {
		done <- writeHumanTrace(pw, summary, detail)
	}()

	collectUntil := func(want string, timeout time.Duration) {
		deadline := time.After(timeout)
		for {
			select {
			case <-outputCh:
				if strings.Contains(allOutput.String(), want) {
					return
				}
			case <-deadline:
				return
			}
		}
	}

	collectUntil("hello", 5*time.Second)
	initText := allOutput.String()
	if !strings.Contains(initText, "⚡ running") {
		t.Fatalf("expected ⚡ running in output, got:\n%s", initText)
	}
	if !strings.Contains(initText, "hello") {
		t.Fatalf("expected 'hello' in output, got:\n%s", initText)
	}

	// Allow time for Watch to fully initialize before appending
	time.Sleep(200 * time.Millisecond)

	newEvents := []string{
		`{"type":"item.completed","item":{"id":"msg_2","type":"agent_message","text":"world"}}`,
	}
	appendLines(t, logPath, newEvents)

	// Allow time for Watch to process the Write event before metadata changes
	time.Sleep(200 * time.Millisecond)

	completedMeta := initialMeta
	completedMeta.Status = "completed"
	completedMeta.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	writeJSON(t, metaPath, completedMeta)

	collectUntil("Done", 10*time.Second)
	followText := allOutput.String()
	if !strings.Contains(followText, "world") {
		t.Fatalf("expected 'world' in output, got:\n%s", followText)
	}
	if !strings.Contains(followText, "Done") {
		t.Fatalf("expected 'Done' in output, got:\n%s", followText)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("writeHumanTrace: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("writeHumanTrace did not return")
	}
}

func TestWriteHumanTracePrintsCompletedAndExits(t *testing.T) {
	dir := t.TempDir()

	metaPath := filepath.Join(dir, "metadata.json")
	logPath := filepath.Join(dir, "events.jsonl")

	meta := trace.AgentTraceMetadata{
		ID:        "test-completed",
		Status:    "completed",
		Command:   "test-cmd",
		LogPath:   logPath,
		CreatedAt: time.Now().Format(time.RFC3339Nano),
		UpdatedAt: time.Now().Format(time.RFC3339Nano),
	}
	writeJSON(t, metaPath, meta)

	events := []string{
		`{"type":"item.completed","item":{"id":"msg_1","type":"agent_message","text":"done"}}`,
	}
	writeLines(t, logPath, events)

	source := trace.NewSessionDirSource(dir)
	summaries, err := source.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	summary := summaries[0]
	detail, err := source.Get(summary.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	var buf bytes.Buffer
	err = writeHumanTrace(&buf, summary, detail)
	if err != nil {
		t.Fatalf("writeHumanTrace: %v", err)
	}

	text := buf.String()
	if !strings.Contains(text, "✓ completed") {
		t.Fatalf("expected ✓ completed, got:\n%s", text)
	}
	if !strings.Contains(text, "✓ Done") {
		t.Fatalf("expected ✓ Done, got:\n%s", text)
	}
	if !strings.Contains(text, "done") {
		t.Fatalf("expected message 'done', got:\n%s", text)
	}
}

func writeJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeLines(t *testing.T, path string, lines []string) {
	t.Helper()
	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func appendLines(t *testing.T, path string, lines []string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	for _, line := range lines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			t.Fatalf("append to %s: %v", path, err)
		}
	}
	if err := f.Sync(); err != nil {
		t.Fatalf("sync %s: %v", path, err)
	}
}
