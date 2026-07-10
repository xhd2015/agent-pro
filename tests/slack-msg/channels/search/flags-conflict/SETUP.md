# Scenario

**Feature**: --exact and --prefix are mutually exclusive

```
Caller -> slack-msg channels search --exact --prefix QUERY -> stderr -> exit 1
```

## Preconditions

- Both flags set together is an error (designer rule: reject, do not prefer one).

## Steps

1. Clear Slack env vars (no API needed).
2. Leaf sets both flags and a QUERY.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
