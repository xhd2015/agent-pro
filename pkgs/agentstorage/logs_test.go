package agentstorage

import (
	"errors"
	"os"
	"testing"
)

func TestAppendAndReadErrorLog(t *testing.T) {
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSession("log-test", SessionMeta{Runner: "codex-tty"}); err != nil {
		t.Fatal(err)
	}
	if err := AppendErrorLog(store.Home(), "log-test", "prompt-inject", errors.New("socket closed")); err != nil {
		t.Fatal(err)
	}
	records, raw, err := ReadLogs(store.Home(), "log-test")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Level != "error" || records[0].Component != "prompt-inject" || records[0].Message != "socket closed" {
		t.Fatalf("unexpected records: %#v", records)
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		t.Fatalf("expected JSONL data, got %q", raw)
	}
	info, err := os.Stat(LogsPath(store.Home(), "log-test"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("logs permissions = %o, want 600", info.Mode().Perm())
	}
}
