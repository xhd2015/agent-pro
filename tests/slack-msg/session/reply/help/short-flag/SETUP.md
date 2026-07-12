# Scenario

**Feature**: session reply -h

```
slack-msg session reply -h -> usage
```

## Steps

1. Args: session reply -h.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"session", "reply", "-h"}
	return nil
}
```
