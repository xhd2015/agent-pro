# Scenario

**Feature**: session info without session id

```
session info (no --session-id / env) -> session id required; exit 1
```

## Steps

1. No session id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"session", "info"}
	return nil
}
```
