## Preconditions
- The grok binary is available in PATH.
- This subtree covers the `Ask()` operation mode of GrokAgent.

## Steps
1. Set `Operation` to `"ask"` to route through the Ask path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = OpAsk
	return nil
}
```
