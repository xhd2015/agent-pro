## Preconditions
- testdata/ contains a directory with no go.mod, multiple subdirs (some gitignored), each with a DOCTest tree.

## Steps
1. Copy testdata/ to a temp directory outside any Go module.
2. Run `doctest test -v ./...`.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    srcTestData := "./testdata"
    tmpTestData := filepath.Join(t.TempDir(), "testdata")
    if err := copyDir(tmpTestData, srcTestData); err != nil {
        t.Fatalf("copy testdata: %v", err)
    }
    req.WorkDir = tmpTestData
    req.Args = []string{"test", "-v", "./..."}
    return nil
}
```
