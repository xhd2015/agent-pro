# Scenario

**Feature**: non-empty transcript is folded into the runner prompt

```
[user, assistant, ...] + new user text -> BuildContinuationPrompt -> prefix + new turn
```

## Preconditions

- At least one prior `message` event with `role=user` exists before the new prompt.

## Steps

1. Leaf `Setup` populates `req.PriorEvents` with a realistic mini transcript.
2. Leaf sets `req.NewPrompt` for the follow-up turn.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if strings.TrimSpace(req.NewPrompt) == "" {
		req.NewPrompt = "what did I ask?"
	}
	return nil
}
```