# Scenario

**Feature**: missing grok binary without fake hook fails clearly

```
# no grok on PATH, no GROK_SHOW_USAGE_COMMAND -> error mentions grok
grok-show-usage -> resolve grok path -> not found
```

## Preconditions

- `PATH` contains no `grok` executable.
- `GROK_SHOW_USAGE_COMMAND` is not set.

## Steps

1. Set `req.SkipFakeCommand = true` and `req.MinimalPATH = true`.
2. Run and assert non-zero exit; stderr mentions `grok`.

## Context

- Exercises production argv resolution when grok is absent.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SkipFakeCommand = true
	req.MinimalPATH = true
	return nil
}
```