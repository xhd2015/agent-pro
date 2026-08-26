# Scenario

**Feature**: multiple grep patterns keep only prompts that contain every pattern

```
# prompts: "fix timeout" / "fix only" / "timeout only"
FilterUserPrompts(Grep=["fix","timeout"]) -> keep only "fix timeout"
```

## Preconditions

- Op `single` with three chrono prompts.
- GrepSet with two patterns; AND on the same `UserPrompt.Text`.

## Steps

1. Write session with three prompts.
2. Set Grep to fix + timeout.
3. Call single (+ filter).

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
	req.Grep = []string{"fix", "timeout"}
	end := atFixed(-5 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterSingle,
		Title:        "grep multi-and",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end,
			"please fix the timeout path",
			"please fix the retry path",
			"timeout alone is not enough",
		),
	})
	return nil
}
```
