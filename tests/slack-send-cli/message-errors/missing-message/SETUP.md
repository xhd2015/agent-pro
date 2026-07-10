# Scenario

**Feature**: flags only without MESSAGE

```
Caller -> slack-send --token --channel (no message) -> message required
```

## Steps

1. Inherit message-errors setup (token + channel flags only).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--token", slackTestToken, "--channel", "C0ALE44K5J6"}
	return nil
}
```