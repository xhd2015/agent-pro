# Scenario

**Feature**: color force keys + TERM policy from injectable parentTERM

```
# force keys always when color
Unset NO_COLOR; Set FORCE_COLOR / CLICOLOR / CLICOLOR_FORCE
# TERM: empty|dumb → xterm-256color; else keep effective parentTERM
```

## Steps

1. Leaves set ParentTERM to dumb / empty / good.
2. Assert shared force keys; leaf-specific TERM expectation.

## Context

- No user env entries here (pure parentTERM path).
- Effective TERM = parentTERM when no `-e TERM=…`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnvEntries = nil
	req.ConfigHome = ""
	req.PrependPaths = nil
	return nil
}

// assertColorForceKeys checks Unset NO_COLOR and the three force assignments.
func assertColorForceKeys(t *testing.T, set, unset []string) {
	t.Helper()
	assertUnsetHas(t, unset, "NO_COLOR")
	assertSetExact(t, set, "FORCE_COLOR", "1")
	assertSetExact(t, set, "CLICOLOR", "1")
	assertSetExact(t, set, "CLICOLOR_FORCE", "1")
}
```
