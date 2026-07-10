# Scenario

**Feature**: auth status help lists flags including --app without API calls

```
Caller -> slack-msg auth status -h|--help -> usage on stdout -> exit 0
```

## Preconditions

- No token required for help path.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h` or `--help` after `auth status`.

## Context

- Help must document `--app`, `--token`, `--app-token`, `--config`, `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
