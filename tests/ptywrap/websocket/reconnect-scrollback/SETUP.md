# Scenario

**Feature**: scrollback replayed when WS client reconnects

```
# disconnect preserves session
first attach -> output -> close WS -> second attach -> scrollback includes marker
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "ws-reconnect"
	return nil
}
```