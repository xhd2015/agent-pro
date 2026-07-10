# Scenario

**Feature**: slack-msg channels list prints visible non-archived channels

```
Caller -> slack-msg channels list [options] -> conversations.list -> sorted human/JSON
```

## Preconditions

- Subcommand path starts with `channels list`.

## Steps

1. Isolate workdir for list leaves.
2. Leaves set `req.Args` starting with `"channels", "list"`.
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
