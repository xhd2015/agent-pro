# Scenario

**Feature**: history API returns error

```
slack-msg history -> conversations.history ok=false -> history failed:
```

## Steps

1. Use history-fail slacktest server.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.HistoryAPIFail = true
	req.Args = []string{
		"history",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
	}
	return nil
}
```
