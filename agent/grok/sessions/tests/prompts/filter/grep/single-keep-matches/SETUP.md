# Scenario

**Feature**: single session with grep keeps only matching user prompts

```
# 3 prompts: alpha / beta-noise / ALPHA-tail
Prompts + FilterUserPrompts(Grep="alpha") -> 2 kept (case-insensitive)
```

## Preconditions

- Session idFilterSingle with three chrono user prompts.
- GrepSet, Grep=`alpha` (matches first and third via case-insensitivity for third may be separate leaf; here both "alpha" substrings).
- Op `single`.

## Steps

1. Write session with three prompts.
2. Set Grep=alpha, GrepSet.
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
	req.Grep = []string{"alpha"}
	end := atFixed(-5 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterSingle,
		Title:        "grep single",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end,
			"alpha one",
			"beta noise only",
			"prefix ALPHA suffix",
		),
	})
	return nil
}
```
