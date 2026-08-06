# Scenario

**Feature**: grep match is case-insensitive literal

```
# prompts: "ERROR boom", "all good", "Error path"
Grep="error" -> keeps first and third
```

## Preconditions

- Single session; Grep=`error` (lowercase); texts use mixed case.
- Op `single`.

## Steps

1. Write three prompts with mixed ERROR casing.
2. Filter with Grep=error.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "single"
	req.SessionID = idFilterSingle
	req.GrepSet = true
	req.Grep = "error"
	end := atFixed(-3 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterSingle,
		Title:        "grep ci",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end,
			"ERROR boom",
			"all good",
			"Error path",
		),
	})
	return nil
}
```
