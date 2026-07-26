# Scenario

**Feature**: help topic `add-missing-scope` prints OAuth scope grant guideline

```
slack-msg --help --topic add-missing-scope
  OR
slack-msg --topic add-missing-scope --help
  -> stdout guideline (groups:read, reinstall, botToken) -> exit 0
```

## Preconditions

- Topic name is exactly `add-missing-scope`.
- Both flag orders must produce the same body.

## Steps

1. Reaffirm no Slack API URL for pure help.
2. Leaf sets args with both `--help` and `--topic add-missing-scope` (order varies).

## Context

- Body is operator-facing prose; ASSERTs lock key phrases via contains, not a
  brittle full golden (except trailing newline + exit 0 + empty stderr).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SlackAPIURL = ""
	return nil
}
```
