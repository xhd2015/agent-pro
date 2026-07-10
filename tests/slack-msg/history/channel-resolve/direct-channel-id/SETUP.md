# Scenario

**Feature**: history with direct channel ID

```
slack-msg history --channel C0ALE44K5J6 -> history OK
```

## Steps

1. Direct channel ID flag.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"history",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
	}
	return nil
}
```
