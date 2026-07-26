## Expected

- `events.jsonl` did not exist before reconcile (`EventLineCountBefore == 0`).
- After reconcile: at least one user message with `reconcile heal probe prompt`.
- Assistant reply `reconcile-heal-reply-marker` present.
- `EventLineCountAfter > EventLineCountBefore`.

## Errors

- Empty events after reconcile indicates `ReconcileOnce` / discovery not implemented.

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.EventLineCountBefore != 0 {
		t.Fatalf("precondition: want 0 events before reconcile, got %d", resp.EventLineCountBefore)
	}
	if resp.EventLineCountAfter <= resp.EventLineCountBefore {
		t.Fatalf("events after reconcile: before=%d after=%d", resp.EventLineCountBefore, resp.EventLineCountAfter)
	}
	if count := countUserMessagesByText(resp.Events, reconcileHealPrompt); count < 1 {
		t.Fatalf("user prompt %q: want >= 1 got %d", reconcileHealPrompt, count)
	}
	foundReply := false
	for _, ev := range resp.Events {
		if ev.Type == types.ActionMessage && ev.Role == "assistant" && strings.Contains(ev.Text, reconcileHealReply) {
			foundReply = true
			break
		}
	}
	if !foundReply {
		t.Fatalf("missing assistant reply %q", reconcileHealReply)
	}
}
```