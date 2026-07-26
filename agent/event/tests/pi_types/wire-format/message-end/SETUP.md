# Scenario

**Feature**: message_end carries a `message` field with the full message

## Preconditions
- message_end carries a `message` field with the full message.

## Steps
1. Parse a message_end event with a complete assistant message.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"Hello world"},{"type":"toolCall","id":"tc_1","name":"bash","arguments":{"command":"ls"}}],"api":"openai-completions","provider":"openai","model":"gpt-4","usage":{"input":10,"output":20,"cacheRead":0,"cacheWrite":0,"totalTokens":30,"cost":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0,"total":0}},"stopReason":"stop","timestamp":1000}}`
	return nil
}
```
