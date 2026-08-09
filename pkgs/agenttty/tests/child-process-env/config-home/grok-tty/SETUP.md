# Scenario

**Feature**: S5 — configHome + grok-tty → GROK_HOME in Set

```
# S5
runnerID=grok-tty, configHome=/tmp/…
  -> Set contains GROK_HOME=/tmp/…
  -> no CODEX_HOME
```

## Steps

1. Set RunnerID to `grok-tty`.
2. Assert GROK_HOME equals ConfigHome.

## Context

- Non-codex-tty runners use GROK_HOME (same as RunnerConfigHomeEnv default branch).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.RunnerID = "grok-tty"
	return nil
}
```
