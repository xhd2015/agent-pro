## Preconditions
- The pi binary is available in PATH.
- This is a fresh session query (no --session-id flag).

## Steps
1. Set the prompt to ask for the capital of France in one word.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "what is the capital of France? answer in one word"
	return nil
}
```
