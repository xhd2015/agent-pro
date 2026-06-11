## Preconditions
- A temporary project with go.mod + multiple DOCTEST.md trees exists.

## Steps
1. Create temp project via parent helper.
2. Run `doctest test ./...` from a subdirectory of the project root.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    projDir, doctestBin := createTempProject(t, req)

    subDir := filepath.Join(projDir, "alpha_test", "simple")
    if mkErr := os.MkdirAll(subDir, 0755); mkErr != nil {
        t.Fatalf("mkdir subDir: %v", mkErr)
    }

    os.Setenv("DOCTEST_BIN", doctestBin)
    req.WorkDir = subDir
    req.Args = []string{"test", "-v", "./..."}
    return nil
}
```
