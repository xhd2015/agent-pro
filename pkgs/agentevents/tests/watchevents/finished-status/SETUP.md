# Scenario

**Bug**: finished session status must not end `WatchEvents` early

```
meta.status == finished at watch start
  -> append while ctx alive must still invoke onLine
```

## Steps

1. Grouping setup documents finished-status branch (no extra mutation).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Finished-status branch: watch starts after UpdateSessionStatus(..., "finished").
	if req.AfterOffset == 0 {
		req.AfterOffset = -1 // sentinel: Run resolves to store.ReadEvents EOF offset
	}
	return nil
}
```