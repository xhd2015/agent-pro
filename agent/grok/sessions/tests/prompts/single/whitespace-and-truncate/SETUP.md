# Scenario

**Feature**: format collapses whitespace and soft-truncates long prompt bodies

```
# structured keeps raw text; format collapses + truncates
user text with tabs/newlines + 250-rune body
  -> Prompts Text raw
  -> FormatPromptsText collapses spaces; body ~200 runes + "…"
```

## Preconditions

- One user prompt whose raw text has internal newlines/tabs and length > 200 runes.
- Op is `format-single` so both structured and Output are available.
- Location UTC.

## Steps

1. Write session with long multi-whitespace user prompt.
2. Call Prompts + FormatPromptsText.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-single"
	req.SessionID = idWhitespace
	// 50 spaces/tabs/newlines mixed prefix + 220 x's so formatted body truncates.
	raw := "hello\t\tworld\n\n" + longPromptRunes(220)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idWhitespace,
		Title:        "whitespace truncate",
		LastActiveAt: atFixed(-10 * time.Minute),
		Updates: updatesJSONL(
			userChunkAt(raw, atFixed(-10*time.Minute)),
			assistantChunk("ack"),
			turnCompleted(),
		),
	})
	return nil
}
```
