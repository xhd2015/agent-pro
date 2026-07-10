# Scenario

**Feature**: help flags print usage without Socket Mode connect

```
Caller -> slack-listen listen -h|--help -> usage on stdout -> exit 0 (no WS connect)
```

## Preconditions

- No tokens required for help path.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h` or `--help` in `req.Args`.

## Context

- Help must reflect `slack-listen listen [options]` contract.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```