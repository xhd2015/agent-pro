# Scenario

**Feature**: flat layout with default-off flags creates no questions or progress dirs

```
# Dir only, zero flags
subagent.Run -> events.jsonl without questions/ or progress/
```

## Preconditions

- Inherited `configureFlatDirFlagsOff` from grouping.

## Steps

1. Inherit flags-off flat session configuration.

## Context

- Inner session id: `inner_flags_off_sess`

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.SessionDir == "" {
		configureFlatDirFlagsOff(t, req)
	}
	return nil
}```
