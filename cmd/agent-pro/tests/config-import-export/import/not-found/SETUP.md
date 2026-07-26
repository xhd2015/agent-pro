## Preconditions
- The zip file specified in ZipPath does not exist.
- The operation has been set to "import".

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Operation = "import"
	req.Agent = "opencode"
	req.ZipPath = filepath.Join(req.HomeDir, "nonexistent.zip")
	return nil
}
```
