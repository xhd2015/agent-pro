# Scenario

**Feature**: UnwrapEvent gracefully handles a non-JSON garbage input string

## Preconditions
- Input is not valid JSON at all (garbage string).
- The function must handle this gracefully.

## Steps
1. Set `SSEInput` to a non-JSON string.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SSEInput = `not valid json at all`
	return nil
}
```
