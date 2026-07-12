# Scenario

**Feature**: session history without session id

```
session history (no --session-id / env) -> session id required; exit 1
```

## Steps

1. No session id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"session", "history"}
	return nil
}
```
