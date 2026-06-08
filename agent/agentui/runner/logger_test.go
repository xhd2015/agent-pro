package runner

import (
	"bytes"
	"testing"

	"github.com/xhd2015/agent-pro/agent_trace/types"
)

func TestFormatLogLineInvalidJSON(t *testing.T) {
	if got := FormatLogLine("not-json"); got != "" {
		t.Fatalf("FormatLogLine invalid = %q, want empty", got)
	}
}

func TestFormatLogLineMessage(t *testing.T) {
	messageLine := `{"type":"text","timestamp":1,"sessionID":"sid_1","part":{"text":"hello"}}`
	if got := FormatLogLine(messageLine); got == "" {
		t.Fatal("expected formatted message log")
	}
}

func TestLoggerBuffersPartialLine(t *testing.T) {
	var file bytes.Buffer
	ch := make(chan string, 4)
	logger := NewLogger(ch, &file)

	logger.Log(`{"type":"text"`)
	if file.Len() != 0 {
		t.Fatalf("partial line wrote %q, want no write", file.String())
	}
	if len(ch) != 0 {
		t.Fatalf("partial line emitted %d messages, want 0", len(ch))
	}
}

func TestLoggerEmitsOnNewlineAndSkipsEmptyLines(t *testing.T) {
	var file bytes.Buffer
	ch := make(chan string, 4)
	logger := NewLogger(ch, &file)

	line := `{"type":"text","timestamp":1,"sessionID":"sid_1","part":{"text":"hello"}}`
	logger.Log("\n")
	logger.Log(line[:20])
	logger.Log(line[20:] + "\n")

	if file.String() != line+"\n" {
		t.Fatalf("file = %q, want %q", file.String(), line+"\n")
	}
	if len(ch) != 1 {
		t.Fatalf("channel messages = %d, want 1", len(ch))
	}
}

func TestLoggerHandlesMultipleLinesInOneChunk(t *testing.T) {
	var file bytes.Buffer
	ch := make(chan string, 4)
	logger := NewLogger(ch, &file)

	line1 := `{"type":"text","timestamp":1,"sessionID":"sid_1","part":{"text":"hello"}}`
	line2 := `{"type":"text","timestamp":2,"sessionID":"sid_1","part":{"text":"world"}}`
	logger.Log(line1 + "\n" + line2 + "\n")

	if len(ch) != 2 {
		t.Fatalf("channel messages = %d, want 2", len(ch))
	}
}

func TestLoggerNonBlockingChannel(t *testing.T) {
	ch := make(chan string, 1)
	ch <- "full"
	logger := NewLogger(ch, nil)

	line := `{"type":"text","timestamp":1,"sessionID":"sid_1","part":{"text":"hello"}}`
	logger.Log(line + "\n")
	if len(ch) != 1 {
		t.Fatalf("channel size changed to %d, want 1", len(ch))
	}
}

func TestPlainLoggerIgnoresEmptyMessages(t *testing.T) {
	var logger PlainLogger
	logger.Log("")
}

func TestTraceAdapterRegisteredForFixtures(t *testing.T) {
	if _, ok := types.ParseAgentTraceLine([]byte(`{"type":"text","timestamp":1,"sessionID":"sid_1","part":{"text":"hello"}}`)); !ok {
		t.Fatal("expected trace adapter registration for fixture")
	}
}
