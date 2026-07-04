# Scenario

**Feature**: send on unknown session id fails when registry is empty

```
# missing registry entry
tty-watch send session-999 "hi" -> error: session not found
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "send-missing"
	req.SendID = "session-99999"
	req.SendMessage = "hi"
	return nil
}
```