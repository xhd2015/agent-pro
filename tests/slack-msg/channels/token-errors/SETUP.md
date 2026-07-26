# Scenario

**Feature**: reject missing bot token for channels

```
Caller -> slack-msg channels list (no token) -> bot token required -> exit 1
```

## Preconditions

- No `--config`, no `SLACK_BOT_TOKEN`.

## Steps

1. Clear Slack env vars.
2. Leaf runs channels list without token.

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
