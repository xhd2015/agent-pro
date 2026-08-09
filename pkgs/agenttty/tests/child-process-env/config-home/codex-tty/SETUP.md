# Scenario

**Feature**: S4 — configHome + codex-tty → CODEX_HOME in Set

```
# S4
runnerID=codex-tty, configHome=/tmp/…
  -> Set contains CODEX_HOME=/tmp/…
  -> no GROK_HOME
```

## Steps

1. Set RunnerID to `codex-tty`.
2. Assert CODEX_HOME equals ConfigHome.

## Context

- Unset stays empty (color false).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.RunnerID = "codex-tty"
	return nil
}
```
