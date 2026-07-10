# Scenario

**Feature**: JSON document with channels sorted by name

```
slack-msg channels list --json --token -> sorted channels array
```

## Steps

1. Flags for token and `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
		"--json",
	}
	return nil
}
```
