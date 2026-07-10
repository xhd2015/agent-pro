# Scenario

**Feature**: --limit 2 yields two name-sorted lines

```
slack-msg channels list --limit 2 -> agent-pro-debug, general
```

## Steps

1. Flags for token and `--limit 2`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
		"--limit", "2",
	}
	return nil
}
```
