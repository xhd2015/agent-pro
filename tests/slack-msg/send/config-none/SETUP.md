# Scenario

**Feature**: stdout reports no config when --config omitted for send

```
Caller -> slack-msg send --token --channel MESSAGE -> Using config from: (none) -> send
```

## Preconditions

- No `--config`, no `SLACK_CONFIG` env.

## Steps

1. Clear Slack env vars.
2. Use CLI flags for token and channel with slacktest backend.

## Context

- `(none)` line is required even when credentials come from CLI flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	apiURL, err := ensureSlackTestServer(t)
	if err != nil {
		return err
	}
	req.SlackAPIURL = apiURL
	return nil
}
```
