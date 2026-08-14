# Scenario

**Feature**: empty occupancy reports input box: empty and JSON empty

```
codex-0.147-empty-glued.txt
  -> DetectInputBox = empty
  -> InputBoxReport human "input box: empty", json "empty"
```

## Preconditions

- Same live empty fixture as the detector leaf.

## Steps

1. Set `req.Fixture` to the empty-glued fixture.

## Context

Locks spelling `input box:` (space, not underscore) vs JSON `input_box`.

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
