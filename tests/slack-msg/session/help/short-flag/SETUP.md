# Scenario

**Feature**: session -h

```
slack-msg session -h -> usage lists reply and history
```

## Steps

1. Args: session -h.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"session", "-h"}
	return nil
}
```
