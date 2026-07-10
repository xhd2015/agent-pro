# Scenario

**Feature**: search without QUERY is rejected

```
Caller -> slack-msg channels search --token TOK -> stderr query required -> exit 1
```

## Preconditions

- No positional QUERY.

## Steps

1. Clear Slack env vars.
2. Leaf provides token only (no QUERY).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
