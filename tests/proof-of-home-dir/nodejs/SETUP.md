## Preconditions
- `node` is available in PATH (test is skipped otherwise).

## Steps
1. Run `node -e 'console.log(require("os").homedir())'` with `HOME` set.
2. The output should be the temporary `HOME` directory path.

```go
import (
    "os/exec"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    if _, err := exec.LookPath("node"); err != nil {
        t.Skipf("skipping: node not found in PATH")
    }
    req.Cmd = "node"
    req.Args = []string{"-e", `console.log(require("os").homedir())`}
    return nil
}
```
