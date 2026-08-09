# Scenario

**Feature**: SubAgent AND Forked — intersection only

```
# subagent_fork kept
# plain subagent dropped (sub, not forked)
# kind=fork dropped (forked, main class)
SubAgent=true + Forked=true
```

## Preconditions

- Both SubAgent and Forked set.

## Steps

1. Write idSub, idSubFork, idFork.
2. SubAgent=true, Forked=true.
3. Only idSubFork remains.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.SubAgent = true
	req.Forked = true

	writeListSessionOpts(t, req.GrokHome, idSub, atFixed(-30*time.Minute), cwdA, "sub not forked", listSessionOpts{
		SessionKind: "subagent",
	})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-20*time.Minute), cwdA, "main fork drop", listSessionOpts{
		SessionKind: "fork",
	})
	writeListSessionOpts(t, req.GrokHome, idSubFork, atFixed(-10*time.Minute), cwdA, "sub forked keep", listSessionOpts{
		SessionKind: "subagent_fork",
	})
	return nil
}
```
