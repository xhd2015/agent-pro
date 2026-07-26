# Scenario

**Feature**: first message in a session has no history prefix

```
(empty prior events) + new user text -> BuildContinuationPrompt -> new prompt only
```

## Preconditions

- `PriorEvents` is nil or empty.

## Steps

1. Leaf leaves `req.PriorEvents` empty.
2. `Run` builds continuation prompt for the new user turn only.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.PriorEvents = nil
	req.NewPrompt = "hi"
	return nil
}
```