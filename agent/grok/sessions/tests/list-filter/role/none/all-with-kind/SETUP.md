# Scenario

**Feature**: no role flags returns every session with Kind tokens populated

```
five sessions: main, sub, sub+, sub-f, fork
  -> ListWithOptions(no MainAgent/SubAgent)
  -> all five newest-first with correct Kind
```

## Preconditions

- Neither role flag set.
- Covers all five display tokens.

## Steps

1. Write five kind fixtures with staggered times.
2. Limit=20; no MainAgent/SubAgent/Forked.
3. Assert IDs + Kind for all five.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20

	// oldest → newest: main, sub, sub+, sub-f, fork
	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-50*time.Minute), cwdA, "main", listSessionOpts{})
	writeListSessionOpts(t, req.GrokHome, idSub, atFixed(-40*time.Minute), cwdA, "sub", listSessionOpts{
		SessionKind: "subagent",
	})
	writeListSessionOpts(t, req.GrokHome, idSubRes, atFixed(-30*time.Minute), cwdA, "sub+", listSessionOpts{
		SessionKind: "subagent_resume",
	})
	writeListSessionOpts(t, req.GrokHome, idSubFork, atFixed(-20*time.Minute), cwdA, "sub-f", listSessionOpts{
		SessionKind: "subagent_fork",
	})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-10*time.Minute), cwdA, "fork", listSessionOpts{
		SessionKind: "fork",
	})
	return nil
}
```
