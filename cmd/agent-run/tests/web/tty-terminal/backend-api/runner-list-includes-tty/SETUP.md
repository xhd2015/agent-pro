# Scenario

**Feature**: web runner catalog includes tty runners

```
GET /api/agent-run/runners -> JSON runners[] includes codex-tty and grok-tty
```

## Preconditions

- Web API runner list is registered.

## Steps

1. Request `/api/agent-run/runners`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.HTTPPath = "/api/agent-run/runners"
	return nil
}
```
