# Scenario

**Feature**: --config supplies token and default channel; message is positional only

```
Caller -> slack-msg send --config PATH MESSAGE -> load JSON -> resolve defaults -> send
```

## Preconditions

- `valid-config.json` or `default-channel-name.json` fixtures.
- `slacktest` for isolated sends.

## Steps

1. Materialize config; insert `--config <abs>` after `send`.
2. Leaf narrows overrides (`--channel`) or default channel shape.
3. Assert success stdout with config path placeholder.

## Context

- CLI `--channel` wins over config `defaultChannelId`.
- Name-shaped `defaultChannelId` resolved like `--channel` names.

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
