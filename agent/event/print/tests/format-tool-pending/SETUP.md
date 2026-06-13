## Preconditions
- An in-progress bash tool event is supplied.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Line = `{"type":"tool_use","part":{"id":"4","type":"tool","tool":"bash","callID":"ca_002","state":{"status":"in_progress","input":{"command":"sleep 10"}}}}`
	return nil
}
```
