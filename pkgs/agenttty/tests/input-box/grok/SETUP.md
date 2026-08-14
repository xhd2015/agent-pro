# Scenario

**Feature**: last live Grok composer glyph is ❯

```
last ❯ (U+2771) line
  -> DetectInputBox
  -> empty if TrimSpace after glyph is empty
  -> occupied if non-empty user text
```

## Preconditions

- Last composer glyph is Grok `❯`. No Grok experiment lock — conservative rule.
- Codex footer glue (` medium · `) does **not** apply to Grok.

## Steps

1. Mark `req.Family=grok` and `ProviderID=grok-tty`.
2. Leaves inject padding-only, user text, or footer-looking text.

## Context

Grok modern idle chrome often pads the `❯` line inside a box border.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Family = "grok"
	req.ProviderID = "grok-tty"
	return nil
}
```
