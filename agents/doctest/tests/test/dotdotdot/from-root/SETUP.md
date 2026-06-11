## Preconditions
- A temporary project with go.mod + multiple DOCTEST.md trees exists.

## Steps
1. Create temp project via parent helper.
2. Run `doctest test ./...` from the project root.

```go
import (
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    projDir, doctestBin := createTempProject(t, req)
    os.Setenv("DOCTEST_BIN", doctestBin)
    req.WorkDir = projDir
    req.Args = []string{"test", "-v", "./..."}
    return nil
}
```
