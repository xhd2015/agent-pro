## Preconditions
- The parent SETUP.md has initialized a temporary home directory.
- This branch tests the `export` operation.

## Steps
1. Mark the operation as export.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "export"
	return nil
}
```
