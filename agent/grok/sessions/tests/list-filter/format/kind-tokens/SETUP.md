# Scenario

**Feature**: FormatListTable renders all five KIND display tokens

```
five sessions: main, sub, sub+, sub-f, fork
  -> ListWithOptions + FormatListTable
  -> Session.Kind set; each token appears as KIND column value on its row
```

## Preconditions

- WantFormat=true.
- One fixture per display token.

## Steps

1. Write five kind fixtures.
2. WantFormat=true, Limit=20.
3. Assert Kind on sessions and each token present on the session's table row.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.WantFormat = true

	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-50*time.Minute), cwdA, "tok-main", listSessionOpts{})
	writeListSessionOpts(t, req.GrokHome, idSub, atFixed(-40*time.Minute), cwdA, "tok-sub", listSessionOpts{
		SessionKind: "subagent",
	})
	writeListSessionOpts(t, req.GrokHome, idSubRes, atFixed(-30*time.Minute), cwdA, "tok-sub+", listSessionOpts{
		SessionKind: "subagent_resume",
	})
	writeListSessionOpts(t, req.GrokHome, idSubFork, atFixed(-20*time.Minute), cwdA, "tok-sub-f", listSessionOpts{
		SessionKind: "subagent_fork",
	})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-10*time.Minute), cwdA, "tok-fork", listSessionOpts{
		SessionKind: "fork",
	})
	return nil
}
```
