## Preconditions
- agent_end carries a `messages` array.

## Steps
1. Parse an agent_end event with a messages array containing assistant and user messages.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "wire"
	req.JSONInput = `{"type":"agent_end","messages":[{"role":"assistant","content":[{"type":"text","text":"Hello"}],"api":"openai-completions","provider":"openai","model":"gpt-4","usage":{"input":10,"output":20,"cacheRead":0,"cacheWrite":0,"totalTokens":30,"cost":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0,"total":0}},"stopReason":"stop","timestamp":1000},{"role":"user","content":"Hi","timestamp":500}]}`
	return nil
}
```
