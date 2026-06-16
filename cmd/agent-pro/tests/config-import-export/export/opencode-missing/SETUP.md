## Preconditions
- No opencode directories exist on disk.
- The operation has been set to "export".

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "export"
	req.Agent = "opencode"
	return nil
}
```
