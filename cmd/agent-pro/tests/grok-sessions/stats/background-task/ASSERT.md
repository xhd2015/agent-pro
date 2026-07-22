## Expected

- `BackgroundTasks` is non-nil.
- Count = 2, AvgMs ≈ 7500, MaxMs ≈ 10000 (wall ms from epoch secs).
- `Subagents` is nil (no subagent_finished lines).

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

	if st.BackgroundTasks == nil {
		t.Fatal("BackgroundTasks is nil")
	}
	if st.BackgroundTasks.Count != 2 {
		t.Fatalf("BackgroundTasks.Count = %d, want 2", st.BackgroundTasks.Count)
	}
	assertFloatNear(t, "BackgroundTasks.AvgMs", st.BackgroundTasks.AvgMs, 7500, 0.01)
	assertFloatNear(t, "BackgroundTasks.MaxMs", st.BackgroundTasks.MaxMs, 10000, 0.01)

	if st.Subagents != nil {
		t.Fatalf("Subagents = %+v, want nil", st.Subagents)
	}
	if !st.Sources.Updates {
		t.Fatal("Sources.Updates = false, want true")
	}
}
```
