## Preconditions
- Source files exist but the zip parent path is a regular file (so `MkdirAll` fails even as root).
- The operation has been set to "export".

## Steps
1. Create opencode config files under home.
2. Set the zip path under a file pretending to be a directory (`not-a-dir/config.zip`).

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

	// Parent must be a file (not chmod-0555 dir): CI containers often run as
	// root, and root can still create files under a 0555 directory. MkdirAll on
	// a file path fails for every uid, which is what Export's create-zip-dir path expects.
	notADir := filepath.Join(req.HomeDir, "not-a-dir")
	if err := os.WriteFile(notADir, []byte("x"), 0644); err != nil {
		t.Fatalf("create not-a-dir file: %v", err)
	}
	req.ZipPath = filepath.Join(notADir, "config.zip")
	return nil
}
```
