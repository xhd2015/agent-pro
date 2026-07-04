# Scenario

**Feature**: send requires both session id and message arguments

```
# missing session id and/or message
tty-watch send -> error
tty-watch send session-1 -> error
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "send-missing-args"
	return nil
}
```