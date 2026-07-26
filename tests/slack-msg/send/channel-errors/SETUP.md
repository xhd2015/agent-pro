# Scenario

**Feature**: reject missing channel when no env/config supplies one

```
Caller -> slack-msg send --token TOK MESSAGE (no channel) -> stderr channel required -> exit 1
```

## Preconditions

- No `--config`, no `SLACK_CHANNEL`.

## Steps

1. Clear Slack env vars.
2. Provide `--token` and MESSAGE only.

## Context

- Channel resolution: CLI → `SLACK_CHANNEL` → config `defaultChannelId`.

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
