package mockpreset

import (
	"os"
	"path/filepath"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func TestLoadJSONL_sleepThenMessage(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ev.jsonl")
	body := "{\"type\":\"sleep\",\"delay_ms\":50}\n{\"type\":\"message\",\"text\":\"done\"}\n"
	if err := os.WriteFile(p, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	evts, err := LoadJSONL(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(evts) != 2 {
		t.Fatalf("len=%d", len(evts))
	}
	if evts[0].Type != types.ActionSleep || evts[0].DelayMs != 50 {
		t.Fatalf("sleep=%+v", evts[0])
	}
	if evts[1].Type != types.ActionMessage || evts[1].Text != "done" {
		t.Fatalf("message=%+v", evts[1])
	}
}
