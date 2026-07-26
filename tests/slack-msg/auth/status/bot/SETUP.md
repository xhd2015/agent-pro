# Scenario

**Feature**: auth status default mode validates bot token via auth.test

```
Caller -> slack-msg auth status [options] -> auth.test -> bot status
```

## Preconditions

- Bot token required unless help path (not under this branch).
- Unit leaves use slacktest `auth.test` fixture with bot_id.

## Steps

1. Clear Slack env by default on unit/success paths as needed.
2. Leaves set args without `--app`.

## Context

- Human field order locked in success leaves.
- Masking: `xoxb-` + `...` + last 4 of full bot token.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
