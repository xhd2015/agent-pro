# Scenario

**Feature**: cursor pages merged then sorted

```
slack-msg channels list --token -> same three lines as single-page fixture
```

## Steps

1. Flags for token only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"channels", "list",
		"--token", slackTestToken,
	}
	return nil
}
```
