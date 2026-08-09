# Scenario

**Feature**: --main-agent keeps main class; drops all sub-agent sessions

```
fixtures: plain main, fork, subagent, subagent_resume, subagent_fork,
          empty-kind+parent, empty-kind-no-parent
  -> ListWithOptions(MainAgent=true)
  -> [empty-no-parent, fork, plain-main] newest-first; Kind main|fork|main
```

## Preconditions

- Seven sessions; timestamps ordered so survivors sort predictably.
- MainAgent=true only role flag.

## Steps

1. Write seven fixtures under cwdA with staggered last_active.
2. MainAgent=true, Limit=20.
3. Expect three survivors: idEmptyNo (newest), idFork, idMain (oldest of survivors).

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.MainAgent = true

	// oldest → newest among main-class survivors: main, fork, empty-no-parent
	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-60*time.Minute), cwdA, "plain main", listSessionOpts{})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-50*time.Minute), cwdA, "fork main-class", listSessionOpts{
		SessionKind: "fork",
	})
	writeListSessionOpts(t, req.GrokHome, idSub, atFixed(-40*time.Minute), cwdA, "subagent drop", listSessionOpts{
		SessionKind: "subagent",
	})
	writeListSessionOpts(t, req.GrokHome, idSubRes, atFixed(-30*time.Minute), cwdA, "sub resume drop", listSessionOpts{
		SessionKind: "subagent_resume",
	})
	writeListSessionOpts(t, req.GrokHome, idSubFork, atFixed(-20*time.Minute), cwdA, "sub fork drop", listSessionOpts{
		SessionKind: "subagent_fork",
	})
	writeListSessionOpts(t, req.GrokHome, idEmptyPar, atFixed(-10*time.Minute), cwdA, "empty+parent drop", listSessionOpts{
		ParentSessionID: idParent,
	})
	writeListSessionOpts(t, req.GrokHome, idEmptyNo, atFixed(-5*time.Minute), cwdA, "empty no parent keep", listSessionOpts{})
	return nil
}
```
