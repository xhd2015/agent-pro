# Scenario

**Feature**: --forked keeps kind=fork, kind=subagent_fork, and forked_at-set sessions

```
three forked fixtures + one plain main (control drop)
  -> ListWithOptions(Forked=true)
  -> three forked kept newest-first
```

## Preconditions

- Forked=true; no role flags.
- forked_at is a non-empty non-whitespace RFC3339-like string.

## Steps

1. Write idFork, idSubFork, idForkedAt (keep) and idMain (drop).
2. Forked=true, Limit=20.
3. Expect idForkedAt, idSubFork, idFork.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.Forked = true

	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-40*time.Minute), cwdA, "plain drop", listSessionOpts{})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-30*time.Minute), cwdA, "kind fork", listSessionOpts{
		SessionKind: "fork",
	})
	writeListSessionOpts(t, req.GrokHome, idSubFork, atFixed(-20*time.Minute), cwdA, "kind subagent_fork", listSessionOpts{
		SessionKind: "subagent_fork",
	})
	writeListSessionOpts(t, req.GrokHome, idForkedAt, atFixed(-10*time.Minute), cwdA, "forked_at only", listSessionOpts{
		ForkedAt: "2026-07-01T12:00:00.000Z",
	})
	return nil
}
```
