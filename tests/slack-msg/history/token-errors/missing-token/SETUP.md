# Scenario

**Feature**: missing bot token for history

```
Caller -> slack-msg history --channel C... -> bot token required
```

## Steps

1. Channel only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"history", "--channel", "C0ALE44K5J6"}
	return nil
}
```
