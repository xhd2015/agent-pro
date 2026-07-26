## Preconditions
- Zip contains entries with a prefix that is not recognized (cursor/).
- The operation has been set to "import".

## Steps
1. Create a zip with unknown prefix entries alongside known ones.
2. Run import.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "import"
	req.Agent = "opencode"
	createZip(t, req.ZipPath, map[string]string{
		"opencode/auth.json":       `{"key":"valid"}`,
		"cursor/settings.json":     `{"cursor":"setting"}`,
	})
	return nil
}
```
