# Scenario

**Feature**: unknown subcommand under channels

```
Caller -> slack-msg channels not-a-command -> stderr + exit 1
```

## Preconditions

- No token required.

## Steps

1. Clear Slack env vars.
2. Leaf uses unknown action under channels.

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
