## Expected

- `resp.Err` is nil (finish-without-spawn does not crash parsing).
- `Subagents.Count == 1`, AvgMs/MaxMs ~ 3000.
- Section `Top subagents` present when TopN default > 0.
- Header includes STATUS / TYPE / TOOLS / TURNS / DESC (or at least not a bare
  crash); TYPE may be empty when spawn missing.
- DESC fallback: a distinctive fragment of `subagent_id` appears
  (`solo-7b0b` or full id). Empty description must not panic.
- Status `completed` and tools `2` may appear when rich columns land.

## Errors

- None (join miss is soft).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	if resp.Stats.Subagents == nil {
		t.Fatal("Subagents is nil")
	}
	if resp.Stats.Subagents.Count != 1 {
		t.Fatalf("Subagents.Count = %d, want 1", resp.Stats.Subagents.Count)
	}
	assertFloatNear(t, "Subagents.AvgMs", resp.Stats.Subagents.AvgMs, 3000, 0.01)
	assertFloatNear(t, "Subagents.MaxMs", resp.Stats.Subagents.MaxMs, 3000, 0.01)

	out := resp.Output
	assertContains(t, out, "Top subagents")
	sub := sectionAfter(out, "Top subagents")
	if sub == "" {
		t.Fatalf("missing Top subagents section:\n%s", out)
	}

	// Id fallback: distinctive fragment of the subagent_id must appear.
	if !strings.Contains(sub, formatTopSubJoinSubID) && !strings.Contains(sub, "solo-7b0b") {
		t.Fatalf("finish-without-spawn: expected id fallback in DESC/label, got:\n%s", sub)
	}
	assertContains(t, sub, "completed")
}
```
