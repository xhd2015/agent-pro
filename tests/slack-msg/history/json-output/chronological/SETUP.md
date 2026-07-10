# Scenario

**Feature**: JSON history chronological messages array

```
slack-msg history --json --token --channel -> {"messages":[...oldest first...],"has_more":false}
```

## Steps

1. Flags for token, channel, and `--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"history",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"--json",
	}
	return nil
}
```
