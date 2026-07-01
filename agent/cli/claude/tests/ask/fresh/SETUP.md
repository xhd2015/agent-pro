# Scenario

**Feature**: Fresh Ask() queries start a new claude session (no resume)

```
# fresh query: no --resume, session_id captured from system init / result
ask/fresh -> ClaudeAgent.Ask(prompt) -> claude -p <prompt> --output-format stream-json --verbose
ClaudeAgent <- claude (system init, assistant blocks, result)
```

## Preconditions
- The `claude` binary is available in PATH.
- This subtree covers fresh session queries (no session resume).

## Steps
1. Ensure no session-resume prompt is set for fresh queries.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Ensure no session-resume prompts are set for fresh queries
	req.ResumePrompt = ""
	return nil
}
```
