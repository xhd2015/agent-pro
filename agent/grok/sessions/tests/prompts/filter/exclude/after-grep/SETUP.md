# Scenario

**Feature**: grep keep then exclude drop (match grep AND not match exclude)

```
# prompts: foo-ok, foobar-bad, other
Grep=foo Exclude=foobar -> only foo-ok
```

## Preconditions

- Both GrepSet and ExcludeSet.
- Op single.

## Steps

1. Write three prompts.
2. Apply grep=foo then exclude=foobar.

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
	req.Grep = []string{"foo"}
	req.ExcludeSet = true
	req.Exclude = "foobar"
	end := atFixed(-4 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterSingle,
		Title:        "grep then exclude",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end,
			"foo-ok",
			"foobar-bad",
			"other",
		),
	})
	return nil
}
```
