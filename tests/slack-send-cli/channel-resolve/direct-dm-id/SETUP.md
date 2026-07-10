# Scenario

**Feature**: DM channel ID used as-is

```
slack-send --channel D024BE91L -> passthrough -> send OK
```

## Steps

1. Pass DM ID directly.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--token", slackTestToken,
		"--channel", "D024BE91L",
		"direct D",
	}
	return nil
}
```