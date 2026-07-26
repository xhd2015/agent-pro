## Preconditions
- Source files exist but the zip path is in a directory that cannot be written to.
- The operation has been set to "export".

## Steps
1. Create opencode config files under home.
2. Set the zip path to a non-existent directory inside a read-only parent.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "export"
	req.Agent = "opencode"
	createSourceFile(t, req.HomeDir, ".local/share/opencode/auth.json", `{"key":"x"}`)
	createSourceFile(t, req.HomeDir, ".config/opencode/opencode.jsonc", `{}`)

	unwritable := filepath.Join(req.HomeDir, "unwritable")
	if err := os.MkdirAll(unwritable, 0555); err != nil {
		t.Fatalf("create unwritable dir: %v", err)
	}
	req.ZipPath = filepath.Join(unwritable, "config.zip")
	return nil
}
```
