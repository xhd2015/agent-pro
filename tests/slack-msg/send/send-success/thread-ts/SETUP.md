# Scenario

**Feature**: send into an existing thread with --thread

```
slack-msg send --token --channel --thread TS MESSAGE -> PostMessage with thread_ts -> OK
```

## Steps

1. Pass `--thread` with a fixed parent timestamp.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"send",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"--thread", "1710000999.000100",
		"thread reply body",
	}
	return nil
}
```
