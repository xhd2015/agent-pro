# Scenario

**Feature**: Grok does not treat Codex footer glue as empty

```
❯ leftover note medium · /tmp
  -> DetectInputBox
  -> occupied
```

## Preconditions

- Last glyph is `❯`, not `›`/`»`.
- Line **contains** ` medium · ` and non-empty user text after the glyph.

## Steps

1. Inject a Grok line that would be empty under the Codex glue rule.

## Context

Keeps Grok conservative: only TrimSpace after `❯`. Applying the Codex footer
rule here would false-empty a user note that happens to mention ` medium · `.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "❯ leftover note medium · /tmp\n"
	req.Fixture = ""
	return nil
}
```
