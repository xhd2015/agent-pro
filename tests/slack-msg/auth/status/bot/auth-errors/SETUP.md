# Scenario

**Feature**: bot auth.test API failure paths

```
slack-msg auth status -> auth.test ok=false -> auth failed: -> exit 1
```

## Preconditions

- Token present so API is called.
- Auth-fail slacktest server (`AuthAPIFail`).

## Steps

1. Leaves set `AuthAPIFail` or rely on parent helpers.
2. Provide `--token`.

## Context

- Stderr prefix `auth failed:`; exit 1.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	return nil
}
```
