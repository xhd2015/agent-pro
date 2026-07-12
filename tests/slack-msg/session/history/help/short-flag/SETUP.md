# Scenario

**Feature**: session history -h

```
slack-msg session history -h -> usage
```

## Steps

1. Args: session history -h.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"session", "history", "-h"}
	return nil
}
```
