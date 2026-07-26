# Scenario

**Feature**: empty transcript does not add a "Previous conversation" block

```
[] + "hi" -> BuildContinuationPrompt -> "hi"
```

## Preconditions

- No prior `message` events.

## Steps

1. Inherited grouping `Setup` sets empty history and `NewPrompt=hi`.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if len(req.PriorEvents) != 0 {
		return fmt.Errorf("first-message leaf: expected empty PriorEvents, got %d", len(req.PriorEvents))
	}
	if req.NewPrompt == "" {
		req.NewPrompt = "hi"
	}
	return nil
}
```