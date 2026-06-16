## Preconditions
- The zip file specified in ZipPath does not exist.
- The operation has been set to "import".

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "import"
	req.Agent = "opencode"
	req.ZipPath = filepath.Join(req.HomeDir, "nonexistent.zip")
	return nil
}
```
