# Scenario

**Feature**: regression guards for existing web/tty-terminal and web/stream contracts

```
websocket-proxy round-trip + resize still works with attach mode
SSE user+assistant contract with CLI-parity shape (no phase)
```

## Preconditions

- Regression leaves mirror prior tree expectations under the refactored backend.

## Steps

1. Grouping setup sets `req.Area = "regression"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Area = "regression"
	return nil
}
```
