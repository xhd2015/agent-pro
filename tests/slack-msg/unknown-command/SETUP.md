# Scenario

**Feature**: unknown subcommand fails with stderr and exit 1

```
Caller -> slack-msg <unknown> -> stderr error -> exit 1
```

## Preconditions

- No network; pure argv validation.

## Steps

1. Clear Slack env vars.
2. Leaf sets unrecognized command name.

## Context

- Unknown command must not be treated as a silent no-op.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
