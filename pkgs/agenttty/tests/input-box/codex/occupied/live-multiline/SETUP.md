# Scenario

**Feature**: live Codex 0.147 multiline draft is occupied

```
› LINE1-DRAFT
  LINE2-DRAFTgpt-5.6-terra medium · /private/…
  -> DetectInputBox
  -> occupied
```

## Preconditions

- Fixture `codex-0.147-occupied-multiline.txt`: first composer line is user
  text without ` medium · `; footer may glue to the wrapped second line.

## Steps

1. Set `req.Fixture` to the occupied-multiline fixture.

## Context

Occupancy is decided on the **last** `›` line (`LINE1-DRAFT`), not the wrap.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Fixture = fixtureCodexOccupiedMultiline
	return nil
}
```
