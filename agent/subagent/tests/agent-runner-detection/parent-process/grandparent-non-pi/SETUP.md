## Preconditions
- The immediate parent (ppid) is not an agent process.
- The grandparent (pppid) is also **not** pi → no agent detected.
- This verifies that the grandparent walk is strictly limited to detecting pi only.

## Steps
1. Each leaf sets `req.ProcessNames` with two entries where the second is not "pi".
2. First hook call returns non-agent ppid name → no match at step 4a.
3. Second hook call returns a non-pi name → no match at step 4b (pi-only check).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    // Grandparent not pi — grandparent walk is pi-only, no match expected
    req.AgentRunnerEnv = ""
    return nil
}
```
