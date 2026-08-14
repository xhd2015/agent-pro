# Scenario

**Feature**: zero-length snapshot is unknown occupancy

```
Scrollback=""
  -> DetectInputBox
  -> unknown
```

## Preconditions

- No fixture file; inline empty string.

## Steps

1. Set `req.Scrollback` to empty.
2. Assert `InputBox=unknown`.

## Context

Same probe outcome as unreachable TTY (no bytes to classify).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = ""
	req.Fixture = ""
	return nil
}
```
