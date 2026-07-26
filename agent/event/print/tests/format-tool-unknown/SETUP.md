## Preconditions
- An event with an unrecognised tool name is supplied.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Line = `{"type":"tool_use","part":{"id":"5","type":"tool","tool":"mystery_tool","callID":"ca_003","state":{"status":"completed","output":"mystery result"}}}`
	return nil
}
```
