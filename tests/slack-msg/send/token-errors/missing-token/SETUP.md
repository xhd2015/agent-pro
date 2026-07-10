# Scenario

**Feature**: missing bot token

```
Caller -> slack-msg send --channel C... MESSAGE -> bot token required
```

## Steps

1. Channel + message only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"send", "--channel", "C0ALE44K5J6", "Hello"}
	return nil
}
```
