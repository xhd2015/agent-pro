# Scenario

**Feature**: grok fresh-session Ask() subtree (no session resume)

## Preconditions
- `Operation` is set to `"ask"` by the parent `ask/` node.
- This subtree covers fresh session queries (no session resume).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Ensure no session-resume prompts are set for fresh queries
	req.ResumePrompt = ""
	return nil
}
```
