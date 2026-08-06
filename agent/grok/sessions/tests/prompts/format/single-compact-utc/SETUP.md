# Scenario

**Feature**: single format emits compact `[UTC timestamp] text` lines with trailing newline

```
# one user prompt at 14:30:00Z
FormatPromptsText -> "[2026-07-03 14:30:00] hello world\n"
```

## Preconditions

- Session with one user prompt text `hello world` at fixedNow−30m.
- Op format-single, Location UTC.

## Steps

1. Write known session.
2. Format via Prompts + FormatPromptsText.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-single"
	req.SessionID = idFormatSingle
	ts := atFixed(-30 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFormatSingle,
		Title:        "format single",
		LastActiveAt: ts,
		Updates:      updatesJSONL(userChunkAt("hello world", ts)),
	})
	return nil
}
```
