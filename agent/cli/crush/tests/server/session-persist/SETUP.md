# Scenario

**Feature**: server-mode CrushAgent Ask resumes a session across two turns

## Preconditions
- Server is running.
- `CrushAgent.Ask()` uses the server client.
- Two calls are made: the first sets up context, the second resumes the session.

## Steps
1. Set `Mode` to `"server-ask"` to use `CrushAgent` with server client.
2. Set prompt and resume prompt.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "server-ask"
	req.Prompt = "what is the capital of France? answer in one word"
	req.ResumePrompt = "what did I ask you about?"
	return nil
}
```
