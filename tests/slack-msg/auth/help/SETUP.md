# Scenario

**Feature**: auth command help lists status without API calls

```
Caller -> slack-msg auth -h|--help -> usage lists status -> exit 0
```

## Preconditions

- No token required for help path.

## Steps

1. Clear Slack env vars.
2. Leaf sets `-h` or `--help` after `auth`.

## Context

- Help must list `status` subcommand.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
