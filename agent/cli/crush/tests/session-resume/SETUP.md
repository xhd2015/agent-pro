## Preconditions
- The crush binary is available in PATH.
- The session-resume test runs two queries: first captures a SessionID, then resumes.

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
