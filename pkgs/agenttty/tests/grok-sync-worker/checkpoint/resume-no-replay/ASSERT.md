## Expected

- Checkpoint exists after stop with `updates_offset` > 0.
- After restart + turn 2: exactly one user message for `resume-turn-one-user`.
- Turn 2 user `resume-turn-two-user` and marker assistant present.
- Total user messages for turn 1 text is 1 (not 2 from replay).

## Errors

- Re-emitting turn 1 after restart indicates checkpoint offset not honored.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureGrokSync: %v", resp.EnsureErr)
	}
	if !resp.CheckpointOK {
		t.Fatal("expected checkpoint after stop")
	}
	if resp.Checkpoint.UpdatesOffset <= 0 {
		t.Fatalf("checkpoint updates_offset must be > 0, got %d", resp.Checkpoint.UpdatesOffset)
	}
	if count := countUserMessagesByText(resp.Events, resumeTurnOneUser); count != 1 {
		t.Fatalf("turn 1 user %q: want 1 (no replay), got %d", resumeTurnOneUser, count)
	}
	if count := countUserMessagesByText(resp.Events, resumeTurnTwoUser); count != 1 {
		t.Fatalf("turn 2 user %q: want 1, got %d", resumeTurnTwoUser, count)
	}
	foundMarker := false
	for _, ev := range resp.Events {
		if strings.Contains(ev.Text, resumeTurnTwoMarker) {
			foundMarker = true
			break
		}
	}
	if !foundMarker {
		t.Fatalf("missing turn 2 marker %q", resumeTurnTwoMarker)
	}
}
```
