# Scenario

**Feature**: slack-msg channels search filters channels by name

```
Caller -> slack-msg channels search [options] QUERY -> list + filter -> human/JSON
```

## Preconditions

- Subcommand path starts with `channels search`.
- QUERY positional required (except help / missing-query leaves).

## Steps

1. Isolate workdir for search leaves.
2. Leaves set `req.Args` starting with `"channels", "search"`.
3. Unit leaves attach slacktest via grouping SETUP.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	return nil
}
```
