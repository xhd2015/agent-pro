# Scenario

**Feature**: Anthropic Messages API — thinking blocks + single breakpoint per HTTP response

```
genQueue dequeue -> messages.go -> content[] with thinking | tool_use | text blocks
```

## Steps

1. Leaves set preset and sequential request bodies (Anthropic messages format).
2. `Assert` parses `content[]` block `type` fields.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Endpoint = "/v1/messages"
	return nil
}
```