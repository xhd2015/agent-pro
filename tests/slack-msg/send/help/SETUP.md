# Scenario

**Feature**: send help flags print usage without sending

```
Caller -> slack-msg send -h|--help -> usage on stdout -> exit 0 (no API call)
```

## Preconditions

- No `--config`, no token/channel/message required for help path.

## Steps

1. Clear Slack env vars so help does not pick up credentials.
2. Leaf sets `-h` or `--help` after `send`.

## Context

- Help must reflect `slack-msg send [options] MESSAGE` including `--thread`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
