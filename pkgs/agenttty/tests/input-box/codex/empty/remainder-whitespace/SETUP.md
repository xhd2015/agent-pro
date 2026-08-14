# Scenario

**Feature**: Codex › with only padding after the glyph is empty

```
"›   \\t  "
  -> DetectInputBox
  -> empty
```

## Preconditions

- Remainder after `›` is whitespace only (TrimSpace empty).
- Line does not need ` medium · `.

## Steps

1. Inject `›` plus spaces/tabs as `req.Scrollback`.

## Context

Second empty rule (complement of footer glue).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "›   \t  \n"
	req.Fixture = ""
	return nil
}
```
