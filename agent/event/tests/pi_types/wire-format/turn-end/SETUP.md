# Scenario

**Feature**: turn_end has `message` and `toolResults` fields

## Preconditions
- turn_end has `message` and `toolResults` fields.

## Steps
1. Parse a turn_end event with assistant message and tool result.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"turn_end","message":{"role":"assistant","content":[{"type":"toolCall","id":"call_1","name":"bash","arguments":{"command":"ls"}}],"api":"openai-completions","provider":"openai","model":"gpt-4","usage":{"input":10,"output":20,"cacheRead":0,"cacheWrite":0,"totalTokens":30,"cost":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0,"total":0}},"stopReason":"toolUse","timestamp":1000},"toolResults":[{"role":"toolResult","toolCallId":"call_1","toolName":"bash","content":[{"type":"text","text":"file.txt"}],"isError":false,"timestamp":1500}]}`
	return nil
}
```
