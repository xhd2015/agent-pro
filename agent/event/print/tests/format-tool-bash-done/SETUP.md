## Preconditions
- A completed bash tool event is supplied.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Line = `{"type":"tool_use","part":{"id":"3","type":"tool","tool":"bash","callID":"ca_001","state":{"status":"completed","output":"hi there","input":{"command":"echo hi"}}}}`
	return nil
}
```
