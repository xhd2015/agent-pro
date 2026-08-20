package agentruncli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func TestRunLogsHumanJSONAndEmpty(t *testing.T) {
	store, err := agentstorage.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSession("with-log", agentstorage.SessionMeta{Runner: "codex-tty"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateSession("empty-log", agentstorage.SessionMeta{Runner: "codex-tty"}); err != nil {
		t.Fatal(err)
	}
	if err := agentstorage.AppendErrorLog(store.Home(), "with-log", "prompt-inject", errors.New("socket closed")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RunLogs([]string{"with-log"}, store, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "error prompt-inject: socket closed") {
		t.Fatalf("human output = %q", out.String())
	}
	out.Reset()
	if err := RunLogs([]string{"--json", "with-log"}, store, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "\"component\":\"prompt-inject\"") {
		t.Fatalf("json output = %q", out.String())
	}
	out.Reset()
	if err := RunLogs([]string{"empty-log"}, store, &out); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "No logs recorded for empty-log.\n"; got != want {
		t.Fatalf("empty output = %q, want %q", got, want)
	}
}
