# Scenario

**Feature**: live empty Codex snapshot stays sendable/idle

```
codex-0.147-empty-glued.txt
  -> DetectInputBox = empty
  -> CheckWritable ready=true state=idle
```

## Preconditions

- Same fixture as `codex/empty/live-placeholder-glued`.

## Steps

1. Set `req.Fixture` to the empty-glued fixture.

## Context

Empty composer ≠ not sendable.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Fixture = fixtureCodexEmptyGlued
	return nil
}
```
