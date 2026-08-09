# Scenario

**Feature**: --sub-agent keeps sub-agent class; drops main-class sessions

```
same multi-kind fixtures as role/main
  -> ListWithOptions(SubAgent=true)
  -> [empty+parent, sub-fork, sub-resume, subagent] newest-first
```

Display Kind note: empty kind + parent is **sub-agent class** for filters,
but KIND column token is still `main` (only explicit session_kind values map to
sub/sub+/sub-f/fork; else → main).

## Preconditions

- SubAgent=true only role flag.
- Empty+parent is sub class with display Kind `main`.

## Steps

1. Write seven fixtures (same matrix as main leaf).
2. SubAgent=true, Limit=20.
3. Expect four survivors: idEmptyPar, idSubFork, idSubRes, idSub.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Limit = 20
	req.SubAgent = true

	writeListSessionOpts(t, req.GrokHome, idMain, atFixed(-60*time.Minute), cwdA, "plain main drop", listSessionOpts{})
	writeListSessionOpts(t, req.GrokHome, idFork, atFixed(-50*time.Minute), cwdA, "fork drop", listSessionOpts{
		SessionKind: "fork",
	})
	writeListSessionOpts(t, req.GrokHome, idSub, atFixed(-40*time.Minute), cwdA, "subagent keep", listSessionOpts{
		SessionKind: "subagent",
	})
	writeListSessionOpts(t, req.GrokHome, idSubRes, atFixed(-30*time.Minute), cwdA, "sub resume keep", listSessionOpts{
		SessionKind: "subagent_resume",
	})
	writeListSessionOpts(t, req.GrokHome, idSubFork, atFixed(-20*time.Minute), cwdA, "sub fork keep", listSessionOpts{
		SessionKind: "subagent_fork",
	})
	writeListSessionOpts(t, req.GrokHome, idEmptyPar, atFixed(-10*time.Minute), cwdA, "empty+parent keep", listSessionOpts{
		ParentSessionID: idParent,
	})
	writeListSessionOpts(t, req.GrokHome, idEmptyNo, atFixed(-5*time.Minute), cwdA, "empty no parent drop", listSessionOpts{})
	return nil
}
```
