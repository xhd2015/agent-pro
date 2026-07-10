# Scenario

**Feature**: channel name/ID resolution reflected in Sending line

```
slack-send --channel CH -> resolve (ID passthrough | knownChannels | conversations.list) -> Sending line shows resolved ID
```

## Preconditions

- `slacktest` with custom `conversations.list` returning `general` → `C0ALE44K5J6`.
- Token via CLI; no `--config` unless leaf overrides.

## Steps

1. Start or reuse slacktest server; set `req.SlackAPIURL`.
2. Leaf sets `--channel` variant.
3. Assert stdout `Sending to channel=RESOLVED` and successful send (proves resolution).

## Context

- Direct `C`/`D`/`G` IDs skip API list lookup.
- `config-known-channels` leaf uses `--config` + knownChannels fast path.

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