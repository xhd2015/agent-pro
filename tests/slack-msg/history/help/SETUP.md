# Scenario

**Feature**: history help flags print usage without API calls

```
Caller -> slack-msg history -h|--help -> usage on stdout -> exit 0
```

## Preconditions

- No token/channel required for help path.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h` or `--help` after `history`.

## Context

- Help must mention `--json`, `--thread`, `--limit`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
