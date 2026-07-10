# Scenario

**Feature**: explicit --config load failures

```
Caller -> slack-send --config PATH ... -> load JSON -> stderr failed to load / empty botToken -> exit 1
```

## Preconditions

- Config is never auto-discovered; only explicit `--config` triggers load.

## Steps

1. Clear Slack env vars.
2. Leaf sets bad config path or empty-token fixture.

## Context

- Bad path fails before send; empty token fails with `botToken is empty in`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```