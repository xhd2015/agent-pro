package events

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenEmptyPath(t *testing.T) {
	logger, err := Open("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger != nil {
		t.Fatalf("expected nil logger for empty path")
	}
}

func TestAppendWritesLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")

	logger, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := logger.Append([]byte(`{"type":"test"}` + "\n")); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := logger.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	expected := `{"type":"test"}` + "\n"
	if string(data) != expected {
		t.Fatalf("got %q, want %q", string(data), expected)
	}
}

func TestMultipleAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")

	logger, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	logger.Append([]byte("line1\n"))
	logger.Append([]byte("line2\n"))
	logger.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "line1\nline2\n" {
		t.Fatalf("got %q, want line1\\nline2\\n", string(data))
	}
}

func TestReopenAppendsDoesNotTruncate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")

	logger1, err := Open(path)
	if err != nil {
		t.Fatalf("Open 1: %v", err)
	}
	logger1.Append([]byte("first\n"))
	logger1.Close()

	logger2, err := Open(path)
	if err != nil {
		t.Fatalf("Open 2: %v", err)
	}
	logger2.Append([]byte("second\n"))
	logger2.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "first\nsecond\n" {
		t.Fatalf("got %q, want first\\nsecond\\n", string(data))
	}
}

func TestCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "events.jsonl")

	logger, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	logger.Append([]byte("data\n"))
	logger.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestAppendOnNilLogger(t *testing.T) {
	var logger *fileLogger
	if err := logger.Append([]byte("data\n")); err != nil {
		t.Fatalf("Append on nil should not error: %v", err)
	}
}

func TestCloseOnNilLogger(t *testing.T) {
	var logger *fileLogger
	if err := logger.Close(); err != nil {
		t.Fatalf("Close on nil should not error: %v", err)
	}
}

func TestOpenNilLoggerOnEmptyPath(t *testing.T) {
	logger, _ := Open("")
	if logger != nil {
		t.Fatalf("expected nil logger")
	}
}

func TestSyncOnNilLogger(t *testing.T) {
	var logger *fileLogger
	if err := logger.Sync(); err != nil {
		t.Fatalf("Sync on nil should not error: %v", err)
	}
}

func TestSyncWrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")

	logger, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	logger.Append([]byte("data\n"))
	if err := logger.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	logger.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "data\n" {
		t.Fatalf("got %q, want data\\n", string(data))
	}
}
