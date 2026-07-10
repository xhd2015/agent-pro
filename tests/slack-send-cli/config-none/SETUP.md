# Scenario

**Feature**: stdout reports no config when --config omitted

```
Caller -> slack-send --token --channel MESSAGE -> Using config from: (none) -> send
```

## Preconditions

- No `--config`, no `SLACK_CONFIG` env.

## Steps

1. Clear Slack env vars.
2. Use CLI flags for token and channel with slacktest backend.

## Context

- `(none)` line is required even when credentials come from CLI flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	apiURL, err := ensureSlackTestServer(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	return nil
}
```