# Scenario

**Feature**: zero / missing wire timestamp formats as `[—]`

```
# user chunk without timestamp field
FormatPromptsText -> line starts with "[—] " then text
```

## Preconditions

- Prefer synthetic SessionPrompts with zero Timestamp so the leaf does not
  depend on convert plumbing for missing ts (still proves format contract).
- Op `format-synthetic`.

## Steps

1. Build synthetic SessionPrompts with one prompt, Timestamp zero, Text `untimed`.
2. FormatPromptsText.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-synthetic"
	req.Synthetic = &sessions.SessionPrompts{
		Session: sessions.Session{
			ID:           idFormatSingle,
			Title:        "missing ts",
			CWD:          fixturePromptsCWD,
			LastActiveAt: fixedNow,
		},
		UserPrompts: []sessions.UserPrompt{
			{Index: 1, Timestamp: time.Time{}, Text: "untimed"},
		},
	}
	return nil
}
```
