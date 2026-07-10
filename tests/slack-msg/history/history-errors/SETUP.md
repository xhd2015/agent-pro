# Scenario

**Feature**: history API and resolution failures

```
slack-msg history -> channel not found or API error -> stderr history failed: -> exit 1
```

## Preconditions

- Default slacktest for channel-not-found; history-fail server for API errors.

## Steps

1. Leaf selects failure mode.
2. Assert `history failed:` prefix and exit 1.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	return nil
}
```
