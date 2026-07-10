# Scenario

**Feature**: missing bot token

```
Caller -> slack-send --channel C... MESSAGE -> bot token required
```

## Steps

1. Inherit token-errors setup.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--channel", "C0ALE44K5J6", "Hello"}
	return nil
}
```