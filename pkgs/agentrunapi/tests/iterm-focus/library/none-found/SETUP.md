# Scenario

**Feature**: resolve succeeds but zero iTerm candidates

```
session ok + (no registry | tree without real TTY | no FindByTTY hit)
  -> FocusSession error "none found"; FocusITerm never
```

## Steps

1. Leaf seeds session; tree/iTerm yield no match.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Grouping: leaves seed session but force zero iTerm candidates.
	req.Phase = "focus"
	req.DryRun = false
	req.Index = nil
	return nil
}
```

