# Scenario

**Feature**: --forked excludes plain main and plain subagent without forked_at

```
plain main + plain subagent + one fork
  -> ListWithOptions(Forked=true)
  -> only fork kept
```

## Preconditions

- Plain subagent has no forked_at.
- Forked=true only.

## Steps

1. Write idMain, idSub, idFork.
2. Forked=true.
3. Only idFork remains.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.Forked = true

	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-30*time.Minute), cwdA, "plain main", listSessionOpts{})
	writeListSessionOpts(t, req.GrokHome, idSub, atFixed(-20*time.Minute), cwdA, "plain subagent", listSessionOpts{
		SessionKind: "subagent",
	})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-10*time.Minute), cwdA, "fork keep", listSessionOpts{
		SessionKind: "fork",
	})
	return nil
}
```
