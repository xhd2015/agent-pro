## Expected
- Exit code 0.
- `$HOME` variable equals the temporary `HOME` directory.
- `~` expands to the temporary `HOME` directory.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil && resp.ExitCode == 0 {
        t.Fatalf("run failed: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
    }

    lines := strings.Split(strings.TrimSpace(resp.Stdout), "\n")
    if len(lines) != 2 {
        t.Fatalf("expected 2 lines (HOME then ~), got %d: %q", len(lines), resp.Stdout)
    }

    home := strings.TrimSpace(lines[0])
    tilde := strings.TrimSpace(lines[1])

    if home != req.TmpHome {
        t.Fatalf("$HOME: expected %q, got %q", req.TmpHome, home)
    }
    if tilde != req.TmpHome {
        t.Fatalf("~: expected %q, got %q", req.TmpHome, tilde)
    }
}
```
