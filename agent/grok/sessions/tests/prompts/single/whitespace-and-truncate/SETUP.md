# Scenario

**Feature**: format collapses whitespace; long body is **full** by default (no soft-cap)

```
# structured keeps raw text; format collapses only (no body-length cap)
user text with tabs/newlines + 220 x runes
  -> Prompts Text raw
  -> FormatPromptsText collapses spaces; full 220 x's; no body-cap "…"
```

## Preconditions

- One user prompt whose raw text has internal newlines/tabs and length > 200 runes.
- Op is `format-single` so both structured and Output are available.
- `!MaxBodySet` (default full body).
- Location UTC.

## Steps

1. Write session with long multi-whitespace user prompt.
2. Call Prompts + Format/WritePromptsText (default opts).

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-single"
	req.SessionID = idWhitespace
	// Tabs/newlines + 220 x's: collapse still; default must NOT soft-cap.
	raw := "hello\t\tworld\n\n" + longPromptRunes(220)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idWhitespace,
		Title:        "whitespace full body",
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
