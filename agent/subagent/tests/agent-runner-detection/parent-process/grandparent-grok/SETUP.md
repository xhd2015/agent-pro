## Preconditions
- The immediate parent (ppid) is **not** an agent process (e.g., bash, zsh).
- The grandparent (pppid) **is** grok → the two-level ancestor walk detects it.
- Grandparent walk is **only applied for pi and grok**, not for opencode/codex/crush.

## Steps
1. Each leaf sets `req.ProcessNames` with two entries: [ppid_name, pppid_name="grok"].
2. First hook call returns ppid_name (non-agent) → no match at step 4a.
3. Second hook call returns "grok" → match at step 4b (pi+grok grandparent walk).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    // Grandparent walk for grok — ppid is non-agent, pppid is "grok"
    req.AgentRunnerEnv = ""
    return nil
}
```
