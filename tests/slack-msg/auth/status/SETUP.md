# Scenario

**Feature**: auth status shows bot or app token status

```
Caller -> slack-msg auth status [--app] [options] -> status human|--json
```

## Preconditions

- Action subcommand is `status` as second arg.

## Steps

1. Leaves set args starting with `"auth", "status"`.
2. Bot unit leaves attach default slacktest; app unit leaves same (connections.open).

## Context

- Default mode validates bot via `auth.test`.
- `--app` validates app via `apps.connections.open`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
