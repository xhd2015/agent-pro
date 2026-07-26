## Preconditions
- `bash` is available in PATH (test is skipped otherwise).

## Steps
1. Run `bash -c 'printf "%s\n%s\n" "$HOME" ~'` with `HOME` set to the fake home directory.
2. Both `$HOME` and `~` expansion should output the fake home directory path.

```go
import (
    "os/exec"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    if _, err := exec.LookPath("bash"); err != nil {
        t.Skipf("skipping: bash not found in PATH")
    }
    req.Cmd = "bash"
    req.Args = []string{"-c", `printf "%s\n%s\n" "$HOME" ~`}
    return nil
}
```
