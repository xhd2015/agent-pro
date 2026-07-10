# Scenario

**Feature**: print thread replies oldest→newest

```
slack-msg history --thread 1710001000.000100 --channel C0... -> parent then replies chronological
```

## Steps

1. Flags for token, channel, and `--thread`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"history",
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"--thread", "1710001000.000100",
	}
	return nil
}
```
