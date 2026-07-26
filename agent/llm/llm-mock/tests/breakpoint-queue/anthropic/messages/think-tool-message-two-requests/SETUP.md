# Scenario

**Feature**: Anthropic encoder — #1 thinking+tool_use, #2 text

```
genQueue [think, tool_call, message]
POST #1 /v1/messages -> thinking block + tool_use bash
POST #2 /v1/messages -> text block with message text
```

## Steps

1. `--mock-events-preset=think-tool-message`.
2. Send two Anthropic messages requests.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.MockEventsPreset = "think-tool-message"
	req.Requests = []string{
		`{"model":"mock-model","max_tokens":1024,"messages":[{"role":"user","content":"anthropic-bp-1"}]}`,
		`{"model":"mock-model","max_tokens":1024,"messages":[{"role":"user","content":"anthropic-bp-2"}]}`,
	}
	return nil
}
```