---
label: e2e
---

## Expected

- `events.jsonl` exists under `AGENT_RUN_HOME/sessions/grok-tty/<id>/`.
- Exactly **one** user `message` with text `hello?`.
- Exactly **one** user `message` with text `what did I say?`.
- At least two `done` events (one per turn).
- Follow-up B HTTP status is accepted (202 or 200).

## Side Effects

- `grok-sync.json` may exist after fix (optional; not asserted here).

## Errors

- Duplicate user rows (differing only by `extensions.grok_session.turn_index`) indicate
  overlapping `startGrokFollowUpEventTail` goroutines — PRIMARY failure mode.

## Exit Code

N/A (HTTP integration probe)

```go
import (
	"net/http"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.FollowUpBStatus != http.StatusAccepted && resp.FollowUpBStatus != http.StatusOK {
		t.Fatalf("follow-up B status=%d body=%q", resp.FollowUpBStatus, resp.FollowUpBBody)
	}
	if resp.EventsFilePath == "" {
		t.Fatal("events.jsonl path empty")
	}
	if resp.UserCountA != 1 {
		t.Fatalf("user count for %q: got %d want 1\n%s",
			req.PromptA, resp.UserCountA, strings.Join(resp.EventsFileLines, "\n"))
	}
	if resp.UserCountB != 1 {
		t.Fatalf("user count for %q: got %d want 1\n%s",
			req.PromptB, resp.UserCountB, strings.Join(resp.EventsFileLines, "\n"))
	}
	if resp.DoneCount < 2 {
		t.Fatalf("done count: got %d want >= 2", resp.DoneCount)
	}
	for _, prompt := range []string{req.PromptA, req.PromptB} {
		indices := turnIndicesForUserPrompt(resp.EventsParsed, prompt)
		if len(indices) > 1 && indices[0] != indices[1] {
			t.Fatalf("duplicate user %q with conflicting turn_index %v (overlapping tails fingerprint)",
				prompt, indices)
		}
	}
}

func turnIndicesForUserPrompt(events []map[string]any, prompt string) []int {
	var out []int
	for _, ev := range events {
		if ev["type"] != "message" || ev["role"] != "user" {
			continue
		}
		text, _ := ev["text"].(string)
		if text != prompt {
			continue
		}
		idx := -1
		if ext, ok := ev["extensions"].(map[string]any); ok {
			if gs, ok := ext["grok_session"].(map[string]any); ok {
				if v, ok := gs["turn_index"].(float64); ok {
					idx = int(v)
				}
			}
		}
		out = append(out, idx)
	}
	return out
}
```
