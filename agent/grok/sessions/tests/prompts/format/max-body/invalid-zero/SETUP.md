# Scenario

**Feature**: MaxBodySet with MaxBodyRunes=0 is invalid (N ≥ 1 required)

```
# MaxBodySet MaxBodyRunes=0
WritePromptsText -> clear error (max-body / >= 1)
```

## Preconditions

- Op format-synthetic (no FS); synthetic has one short prompt so format would
  otherwise succeed.
- MaxBodySet true; MaxBodyRunes=0.
- Package validation surfaces via Write* error (harness sets resp.Err).

## Steps

1. Build synthetic SessionPrompts with one prompt.
2. Format with invalid MaxBody 0.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-synthetic"
	req.MaxBodySet = true
	req.MaxBodyRunes = 0
	req.Synthetic = &sessions.SessionPrompts{
		UserPrompts: []sessions.UserPrompt{
			{Index: 1, Timestamp: atFixed(-time.Minute), Text: "short"},
		},
	}
	return nil
}
```
