## Preconditions
- The parent SETUP.md has initialized a temporary home directory.
- This branch tests the `import` operation.

## Steps
1. Mark the operation as import.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "import"
	return nil
}
```
