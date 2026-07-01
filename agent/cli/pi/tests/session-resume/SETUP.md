# Scenario

**Feature**: PiAgent two-turn session resume reusing the captured SessionID

## Preconditions
- The pi binary is available in PATH.
- Session resume runs two queries: first captures SessionID from the session header line, second reuses it via `--session-id`.

## Steps
1. Set the initial prompt to ask for the capital of France in one word.
2. Set the resume prompt to ask what was previously asked.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "what is the capital of France? answer in one word"
	req.ResumePrompt = "what did I ask you about?"
	return nil
}
```
