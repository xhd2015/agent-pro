# Scenario

**Feature**: unrecognized command name

```
Caller -> slack-msg not-a-command -> stderr + exit 1
```

## Steps

1. Args `["not-a-command"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"not-a-command"}
	return nil
}
```
