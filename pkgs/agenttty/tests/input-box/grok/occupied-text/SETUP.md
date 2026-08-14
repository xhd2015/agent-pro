# Scenario

**Feature**: Grok ❯ with user text is occupied

```
❯ leftover note
  -> DetectInputBox
  -> occupied
```

## Preconditions

- Remainder after last `❯` is non-empty after TrimSpace.

## Steps

1. Inject `❯ leftover note`.

## Context

Suggested scenario 8 from the design requirement.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "❯ leftover note\n"
	req.Fixture = ""
	return nil
}
```
