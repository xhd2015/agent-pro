## Expected

- Turn 2 user message `turn-index-restore-user-prompt` has `turn_index: 1`.
- Turn 0 user message is **not** re-emitted (resume from EOF offset).
- Checkpoint `turn_index` advances to 2 after second `turn_completed`.

## Errors

- `turn_index: 0` on turn 2 user event indicates converter not restored from checkpoint.

## Exit Code

N/A (direct package call)

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureGrokSync: %v", resp.EnsureErr)
	}
	if count := countUserMessagesByText(resp.Events, "turn-zero-user"); count != 0 {
		t.Fatalf("turn 0 user must not replay on resume; got %d", count)
	}
	var turn2User *int
	for _, ev := range resp.Events {
		if ev.Type == types.ActionMessage && ev.Role == "user" && ev.Text == turnIndexRestoreUser {
			idx := grokTurnIndex(ev)
			turn2User = &idx
			break
		}
	}
	if turn2User == nil {
		t.Fatalf("missing turn 2 user %q; events=%d", turnIndexRestoreUser, len(resp.Events))
	}
	if *turn2User != 1 {
		t.Fatalf("turn 2 user turn_index: got %d want 1", *turn2User)
	}
	if resp.Checkpoint.TurnIndex < 2 {
		t.Fatalf("checkpoint turn_index after turn 2 done: got %d want >= 2", resp.Checkpoint.TurnIndex)
	}
}
```
