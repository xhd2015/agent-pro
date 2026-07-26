# Scenario

**Feature**: fresh grok Ask() with an explicit model override (grok-build)

## Preconditions
- The grok binary is available in PATH.
- This is a fresh session query with an explicit model override (`grok-build`).

## Steps
1. Set the prompt to ask for the capital of France in one word.
2. Set the model to `"grok-build"`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "what is the capital of France? answer in one word"
	req.Model = "grok-build"
	return nil
}
```
