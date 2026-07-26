# Scenario

**Feature**: Fresh Ask() with an explicit --model override still returns an answer

```
# model-override: claude invoked with --model haiku
model-override -> ClaudeAgent.Ask(prompt, Model="haiku")
ClaudeAgent -> claude -p <prompt> --output-format stream-json --verbose --model haiku
ClaudeAgent <- claude (assistant text, result)
```

## Preconditions
- The `claude` binary is available in PATH.
- This is a fresh session query with an explicit model override (`haiku`).

## Steps
1. Set the prompt to ask claude to reply with exactly the word "pong".
2. Set the model to `"haiku"`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "Reply with exactly the word: pong"
	req.Model = "haiku"
	return nil
}
```
