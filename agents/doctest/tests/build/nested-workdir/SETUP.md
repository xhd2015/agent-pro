## Preconditions
- A valid doc-style test tree exists in the repository.
- The command is invoked from inside a nested repository directory.

## Steps
1. Set the process working directory to `agents/test-case-tree-runner`.
2. Run `doctest build <absolute-dir>`.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    repoRoot := req.WorkDir
    exampleDir := filepath.Join(repoRoot, "agents/test-case-tree-runner/examples/basic-request-runner")
    req.WorkDir = filepath.Join(repoRoot, "agents/test-case-tree-runner")
    req.Args = []string{"build", exampleDir}
    return nil
}
```
