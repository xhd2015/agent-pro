# Scenario

**Feature**: invalid channel ID triggers send failure

```
slack-send INVALID_CHANNEL -> send failed -> exit 1
```

## Steps

1. Args `["INVALID_CHANNEL", "test"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"INVALID_CHANNEL", "test"}
	return nil
}
```