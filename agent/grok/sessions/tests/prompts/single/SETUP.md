# Scenario

**Feature**: Prompts(sessionID) returns all user prompts for one session

```
# single-session path
writePromptSession(id, updates.jsonl)
  -> Prompts(grokHome, id)
  -> SessionPrompts | not-found error
```

## Preconditions

- Default Op is `single` unless a leaf switches to format-single.
- Unknown / empty id → error containing `grok session not found`.
- Assistant/tool-only or missing updates → empty `UserPrompts`, **no** error.

## Steps

1. Leaf seeds session (or leaves home empty for unknown).
2. Set SessionID and Op.
3. Assert structured prompts or error.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Op == "" {
		req.Op = "single"
	}
	return nil
}
```
