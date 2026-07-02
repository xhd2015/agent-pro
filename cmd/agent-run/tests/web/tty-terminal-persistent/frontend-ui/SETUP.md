# Scenario

**Bug**: chat UI must show terminal affordance from terminal availability, not turn status

```
finished tty chat + terminal status available -> chat top bar terminal button visible
```

## Preconditions

- Browser leaf requires `playwright-debug`.
- Session is `finished`.
- The mapped PTY is live.

## Steps

1. Descendant setup writes mapped finished session and live registry.
2. Browser opens the finished chat route.
3. Browser waits for terminal button.

## Context

- This is the only UI automation branch in the follow-up tree.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "ui"
	req.Status = "finished"
	return nil
}
```
