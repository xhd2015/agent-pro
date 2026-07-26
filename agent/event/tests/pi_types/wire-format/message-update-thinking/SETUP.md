# Scenario

**Feature**: thinking_delta events include `contentIndex`, `delta`, and `partial` message

## Preconditions
- thinking_delta events include `contentIndex`, `delta`, and `partial` message.

## Steps
1. Parse a message_update with an assistant thinking_delta event.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"message_update","message":{"role":"assistant","content":[{"type":"thinking","thinking":"thinking"}],"api":"openai-completions","provider":"openai","model":"gpt-4","usage":{"input":10,"output":5,"cacheRead":0,"cacheWrite":0,"totalTokens":15,"cost":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0,"total":0}},"stopReason":"stop","timestamp":1000},"assistantMessageEvent":{"type":"thinking_delta","contentIndex":0,"delta":" deeper","partial":{"role":"assistant","content":[{"type":"thinking","thinking":"thinking deeper"}],"api":"openai-completions","provider":"openai","model":"gpt-4","usage":{"input":10,"output":10,"cacheRead":0,"cacheWrite":0,"totalTokens":20,"cost":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0,"total":0}},"stopReason":"stop","timestamp":1000}}}`
	return nil
}
```
