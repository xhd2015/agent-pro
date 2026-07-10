# Scenario

**Feature**: help flags print usage without loading config or sending

```
Caller -> slack-send -h|--help -> usage on stdout -> exit 0 (no API call)
```

## Preconditions

- No `--config`, no token/channel/message required for help path.

## Steps

1. Clear Slack env vars so help does not pick up credentials.
2. Leaf sets `-h` or `--help` in `req.Args`.

## Context

- Help must reflect new `slack-send [options] MESSAGE` contract (no auto-config).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```