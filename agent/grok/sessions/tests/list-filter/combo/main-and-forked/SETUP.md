# Scenario

**Feature**: MainAgent AND Forked — intersection only

```
# fork (main + forked) kept
# plain main (main, not forked) dropped
# subagent_fork (forked, sub class) dropped
MainAgent=true + Forked=true
```

## Preconditions

- Both MainAgent and Forked set.
- Three fixtures covering the AND matrix cells.

## Steps

1. Write idMain, idFork, idSubFork.
2. MainAgent=true, Forked=true.
3. Only idFork remains.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.MainAgent = true
	req.Forked = true

	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-30*time.Minute), cwdA, "main not forked", listSessionOpts{})
	writeListSessionOpts(t, req.GrokHome, idSubFork, atFixed(-20*time.Minute), cwdA, "sub forked drop", listSessionOpts{
		SessionKind: "subagent_fork",
	})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-10*time.Minute), cwdA, "main forked keep", listSessionOpts{
		SessionKind: "fork",
	})
	return nil
}
```
