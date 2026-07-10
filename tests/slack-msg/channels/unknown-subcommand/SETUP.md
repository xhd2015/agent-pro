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
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
