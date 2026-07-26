# Scenario

**Feature**: missing session classifies as run

```
# no sessions under store home
Classify(store, "sess-missing", nil)
  -> ModeRun, found=false, err=nil
```

## Preconditions

- Session id not present in store.
- Probe may be nil (unused when not found).

## Steps

1. Use unknown SessionID; do not seed meta.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-missing-never-created"
	req.SeedMeta = false
	req.UseProbe = false
	return nil
}
```
