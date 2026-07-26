# Scenario

**Feature**: reject missing app token for auth status --app

```
Caller -> slack-msg auth status --app (no app token) -> app token required -> exit 1
```

## Preconditions

- No `--app-token`, no `SLACK_APP_TOKEN`, no config appToken.

## Steps

1. Clear Slack env vars.
2. Provide `auth status --app` only.

## Context

- App token resolution: CLI → `SLACK_APP_TOKEN` → config `appToken`.

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
