## Preconditions
- Zip contains an entry with `..` attempting path traversal.
- The operation has been set to "import".

## Steps
1. Create a zip with a `../` path entry.
2. Run import.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "import"
	req.Agent = "opencode"
	createZip(t, req.ZipPath, map[string]string{
		"opencode/auth.json":                    `{"key":"safe"}`,
		"opencode/../../etc/passwd":             `root:x:0:0:root`,
	})
	return nil
}
```
