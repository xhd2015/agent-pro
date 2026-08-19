# Scenario

**Feature**: web agent runs emit phased assistant `message` events for inline streaming UX

```
agentui.Run onDelta -> events.jsonl assistant message phase start|update|end
```

## Preconditions

- `fake-codex` produces streamable assistant output during web create-session run.
- Events persist to `AGENT_RUN_HOME/sessions/.../events.jsonl`.

## Steps

1. Leaf POSTs create session and waits for run completion (or polls events file during run).
2. `Assert` inspects persisted or API events for assistant phases.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "web"
	return nil
}

func readEventsJSONL(t *testing.T, home, runner, sessionID string) []map[string]any {
	t.Helper()
	path := filepath.Join(home, "sessions", sessionID, "events.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	var out []map[string]any
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("invalid event line: %v\n%s", err, line)
		}
		out = append(out, obj)
	}
	return out
}

func waitForAssistantPhases(t *testing.T, home, runner, sessionID string, timeout time.Duration) []map[string]any {
	t.Helper()
	want := map[string]bool{"start": false, "update": false, "end": false}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		events := readEventsJSONL(t, home, runner, sessionID)
		for _, ev := range events {
			if ev["type"] != "message" || ev["role"] != "assistant" {
				continue
			}
			phase, _ := ev["phase"].(string)
			if phase != "" {
				want[phase] = true
			}
		}
		if want["start"] && want["update"] && want["end"] {
			return readEventsJSONL(t, home, runner, sessionID)
		}
		time.Sleep(50 * time.Millisecond)
	}
	events := readEventsJSONL(t, home, runner, sessionID)
	t.Fatalf("timeout waiting for assistant phases start/update/end, last events: %v", events)
	return events
}

func assistantStreamIDs(events []map[string]any) map[string]bool {
	ids := map[string]bool{}
	for _, ev := range events {
		if ev["type"] != "message" || ev["role"] != "assistant" {
			continue
		}
		phase, _ := ev["phase"].(string)
		if phase == "" {
			continue
		}
		id, _ := ev["id"].(string)
		if id != "" {
			ids[id] = true
		}
	}
	return ids
}
```