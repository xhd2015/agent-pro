# Scenario

**Feature**: malformed session positional (no slash) is rejected

```
agent-run sessions notavalidslash --print -> exit 1
```

## Preconditions

- Positional must be exactly one `runner/session_id` token.

## Steps

1. Run print with invalid positional `notavalidslash`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"sessions", "notavalidslash", "--print"}
	return nil
}
```