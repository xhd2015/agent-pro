package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAskAgentCLIUsesFakeCodex(t *testing.T) {
	fakeCodex := buildFakeCodex(t)
	scriptPath := writeRunnerScript(t, `{
  "events": [
    {
      "type": "item.completed",
      "item": {
        "id": "m1",
        "type": "message",
        "text": "fake codex answered",
        "status": "completed"
      }
    }
  ]
}`)
	t.Setenv("AGENT_RUNNER_FAKE_CODEX_PATH", fakeCodex)
	t.Setenv("FAKE_CODEX_SCRIPT", scriptPath)

	logCh := make(chan string, 4)
	sessionDir := t.TempDir()
	output, err := askAgentCLI("answer this", "", "", sessionDir, "fake-codex", logCh)
	if err != nil {
		t.Fatalf("askAgentCLI: %v", err)
	}
	if output != "fake codex answered" {
		t.Fatalf("output = %q, want fake codex answered", output)
	}
	if got := drainLogs(logCh); !strings.Contains(got, "fake codex answered") {
		t.Fatalf("logs = %q, want fake codex answered", got)
	}
	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("read events log: %v", err)
	}
	if !strings.Contains(string(data), "fake codex answered") {
		t.Fatalf("events log missing response:\n%s", string(data))
	}
}

func TestAskAgentCLISurfacesFakeCodexError(t *testing.T) {
	fakeCodex := buildFakeCodex(t)
	scriptPath := writeRunnerScript(t, `{
  "events": [
    {
      "type": "item.completed",
      "item": {
        "id": "m1",
        "type": "message",
        "text": "partial answer",
        "status": "completed"
      }
    }
  ],
  "exit_code": 5,
  "stderr": "script failed"
}`)
	t.Setenv("AGENT_RUNNER_FAKE_CODEX_PATH", fakeCodex)
	t.Setenv("FAKE_CODEX_SCRIPT", scriptPath)

	output, err := askAgentCLI("answer this", "", "", t.TempDir(), "fake-codex", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if output != "partial answer" {
		t.Fatalf("output = %q, want partial answer", output)
	}
	if !strings.Contains(err.Error(), "script failed") {
		t.Fatalf("error = %v, want script failed", err)
	}
}

func buildFakeCodex(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "fake-codex")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", bin, "../../../cmd/fake-codex")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake-codex: %v\n%s", err, string(out))
	}
	return bin
}

func writeRunnerScript(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "script.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return path
}

func drainLogs(ch <-chan string) string {
	var logs []string
	for {
		select {
		case msg := <-ch:
			logs = append(logs, msg)
		default:
			return strings.Join(logs, "\n")
		}
	}
}
