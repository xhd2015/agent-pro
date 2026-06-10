package fakeagent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadScriptSuccess(t *testing.T) {
	path := writeScript(t, `{
  "events": [
    {
      "type": "item.completed",
      "item": {
        "id": "m1",
        "type": "message",
        "text": "scripted done",
        "status": "completed"
      }
    }
  ],
  "exit_code": 3,
  "stderr": "scripted stderr",
  "delay_ms": 12
}`)

	script, err := LoadScript(path)
	if err != nil {
		t.Fatalf("LoadScript: %v", err)
	}
	if script.ExitCode != 3 {
		t.Fatalf("ExitCode = %d, want 3", script.ExitCode)
	}
	if script.Stderr != "scripted stderr" {
		t.Fatalf("Stderr = %q", script.Stderr)
	}
	if script.DelayMS != 12 {
		t.Fatalf("DelayMS = %d, want 12", script.DelayMS)
	}
	if len(script.Events) != 1 {
		t.Fatalf("events len = %d, want 1", len(script.Events))
	}
	if script.Events[0].Item == nil || script.Events[0].Item.Text != "scripted done" {
		t.Fatalf("unexpected message event: %#v", script.Events[0])
	}
}

func TestLoadScriptRejectsInvalidJSON(t *testing.T) {
	path := writeScript(t, `{`)
	_, err := LoadScript(path)
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
	if !strings.Contains(err.Error(), "parse script") {
		t.Fatalf("error = %v, want parse script", err)
	}
}

func TestLoadScriptRejectsMissingEventType(t *testing.T) {
	path := writeScript(t, `{"events":[{"item":{"type":"message"}}]}`)
	_, err := LoadScript(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "events[0].type is required") {
		t.Fatalf("error = %v", err)
	}
}

func writeScript(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "script.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}
