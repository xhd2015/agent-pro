# Scenario

**Feature**: unreachable TTY / no snapshot reports unknown

```
req.Unreachable=true  (scrollback unavailable)
  -> DetectInputBox("") = unknown
  -> InputBoxReport human "input box: unknown", json "unknown"
```

## Preconditions

- Models `tty status` / `status` when TCP is down or `SnapshotText` fails.
- A leftover fixture/scrollback on the request must be ignored.

## Steps

1. Set `req.Unreachable=true` and a non-empty fixture that would otherwise be
   `empty` — reachability wins.

## Context

Unreachable / no snapshot → `unknown`, not “assume empty composer”.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Unreachable = true
	req.Fixture = fixtureCodexEmptyGlued
	return nil
}
```
