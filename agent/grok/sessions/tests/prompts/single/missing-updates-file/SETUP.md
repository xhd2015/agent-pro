# Scenario

**Feature**: summary-only session (no updates.jsonl) yields empty prompts

```
# summary.json present; updates.jsonl omitted
Prompts -> SessionPrompts with UserPrompts empty, Err nil
```

## Preconditions

- Session is findable via summary.json.
- `OmitUpdates` true — no updates.jsonl file.

## Steps

1. Write summary-only session.
2. Call Prompts.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "single"
	req.SessionID = idMissingUpdates
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idMissingUpdates,
		Title:        "no updates file",
		LastActiveAt: atFixed(-2 * time.Hour),
		OmitUpdates:  true,
	})
	return nil
}
```
