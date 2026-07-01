# Scenario

**Feature**: fresh grok Ask() query with no model override or session resume

## Preconditions
- The grok binary is available in PATH.
- This is a fresh session query (no model override, no session-resume).

## Steps
1. Set the prompt to ask for the capital of France in one word.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "what is the capital of France? answer in one word"
	return nil
}
```
