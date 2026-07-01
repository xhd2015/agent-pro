# Scenario

**Feature**: message_start carries a `message` field

## Preconditions
- message_start carries a `message` field.

## Steps
1. Parse a message_start event with a user message containing text content.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"message_start","message":{"role":"user","content":"Hello world","timestamp":1000}}`
	return nil
}
```
