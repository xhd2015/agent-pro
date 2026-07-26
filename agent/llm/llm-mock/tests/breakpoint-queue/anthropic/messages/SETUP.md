# Scenario

**Feature**: Anthropic Messages API — thinking blocks + single breakpoint per HTTP response

```
genQueue dequeue -> messages.go -> content[] with thinking | tool_use | text blocks
```

## Steps

1. Leaves set preset and sequential request bodies (Anthropic messages format).
2. `Assert` parses `content[]` block `type` fields.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Endpoint = "/v1/messages"
	return nil
}
```