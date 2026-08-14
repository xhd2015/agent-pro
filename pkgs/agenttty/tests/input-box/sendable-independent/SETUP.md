# Scenario

**Feature**: occupancy must not change CheckWritable / sendable

```
live empty or occupied Codex snapshot
  -> DetectInputBox (empty|occupied)
  -> CheckWritable still ready=true state=idle
```

## Preconditions

- Uses the locked 0.147 fixtures (same bytes as the detect leaves).
- Experiment: `tty status` was `sendable: yes` / `sendable_state=idle` in both
  empty and occupied.

## Steps

1. Mark `req.Family=sendable` and `ProviderID=codex-tty`.
2. Leaf loads empty-glued or occupied-single fixture.

## Context

Do not overload `sendable` or `screen_status`. Occupancy is a new field.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Family = "sendable"
	req.ProviderID = "codex-tty"
	return nil
}
```
