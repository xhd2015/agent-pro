## Expected

- `Subagents` is non-nil.
- Count = 2, AvgMs ≈ 2000, MaxMs ≈ 3000.
- `BackgroundTasks` is nil (no task_completed lines).

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

	if st.Subagents == nil {
		t.Fatal("Subagents is nil")
	}
	if st.Subagents.Count != 2 {
		t.Fatalf("Subagents.Count = %d, want 2", st.Subagents.Count)
	}
	assertFloatNear(t, "Subagents.AvgMs", st.Subagents.AvgMs, 2000, 0.01)
	assertFloatNear(t, "Subagents.MaxMs", st.Subagents.MaxMs, 3000, 0.01)

	if st.BackgroundTasks != nil {
		t.Fatalf("BackgroundTasks = %+v, want nil", st.BackgroundTasks)
	}
	if !st.Sources.Updates {
		t.Fatal("Sources.Updates = false, want true")
	}
}
```
