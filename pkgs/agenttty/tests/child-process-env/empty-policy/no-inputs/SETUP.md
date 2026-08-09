# Scenario

**Feature**: S1 — no color, no home, no `-e`, no prepend → empty Set/Unset

```
# S1 identity
BuildChildProcessEnv("grok-tty", "", nil, nil, false, "")
  -> Set empty, Unset empty
```

## Steps

1. Explicitly clear all policy inputs (leaf is the identity case).
2. Assert Set and Unset are empty slices.

## Context

- ParentTERM empty must not inject TERM when color is false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Explicit S1 inputs (identity): all optional policy knobs off/empty.
	req.Mode = "build"
	req.RunnerID = "grok-tty"
	req.ConfigHome = ""
	req.PrependPaths = nil
	req.EnvEntries = nil
	req.Color = false
	req.ParentTERM = ""
	return nil
}
```
