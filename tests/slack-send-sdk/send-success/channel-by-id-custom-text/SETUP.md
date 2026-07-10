# Scenario

**Feature**: channel ID and custom message text

```
slack-send C0ALE44K5J6 "custom message here" -> send -> OK
```

## Steps

1. Args `["C0ALE44K5J6", "custom message here"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"C0ALE44K5J6", "custom message here"}
	return nil
}
```