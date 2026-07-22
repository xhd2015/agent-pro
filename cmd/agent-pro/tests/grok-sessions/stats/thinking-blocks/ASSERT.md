## Expected

- `ThinkingBlocks` is 2: one coalesced run of three consecutive thought chunks
  (including nested envelope), then a second run after a non-thought gap.
- `Sources.Updates` is true.
- BackgroundTasks and Subagents remain nil (no task/subagent updates).

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	st := resp.Stats

	if st.ThinkingBlocks != 2 {
		t.Fatalf("ThinkingBlocks = %d, want 2 (coalesced consecutive + gap)", st.ThinkingBlocks)
	}
	if !st.Sources.Updates {
		t.Fatal("Sources.Updates = false, want true")
	}
	if st.BackgroundTasks != nil {
		t.Fatalf("BackgroundTasks = %+v, want nil", st.BackgroundTasks)
	}
	if st.Subagents != nil {
		t.Fatalf("Subagents = %+v, want nil", st.Subagents)
	}
}
```
