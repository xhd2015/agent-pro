## Preconditions
- The file at ZipPath exists but is not a valid zip archive.
- The operation has been set to "import".

## Steps
1. Write a plain text file at the zip path.
2. Run import.

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Operation = "import"
	req.Agent = "opencode"
	if err := os.WriteFile(req.ZipPath, []byte("this is not a zip file"), 0644); err != nil {
		t.Fatalf("write non-zip: %v", err)
	}
	return nil
}
```
