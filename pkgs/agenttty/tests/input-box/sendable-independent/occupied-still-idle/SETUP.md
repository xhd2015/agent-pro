# Scenario

**Feature**: live occupied Codex snapshot stays sendable/idle

```
codex-0.147-occupied-single.txt
  -> DetectInputBox = occupied
  -> CheckWritable ready=true state=idle
```

## Preconditions

- Same fixture as `codex/occupied/live-single-line`.

## Steps

1. Set `req.Fixture` to the occupied-single fixture.

## Context

Occupied composer ≠ not sendable.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Fixture = fixtureCodexOccupiedSingle
	return nil
}
```
