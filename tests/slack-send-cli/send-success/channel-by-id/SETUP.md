# Scenario

**Feature**: send with channel ID flag

```
slack-send --token --channel C0ALE44K5J6 MESSAGE -> OK
```

## Steps

1. Channel ID instead of name.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--token", slackTestToken,
		"--channel", "C0ALE44K5J6",
		"custom text here",
	}
	return nil
}
```