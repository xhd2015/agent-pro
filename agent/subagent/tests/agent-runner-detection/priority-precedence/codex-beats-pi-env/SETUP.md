## Preconditions
- No `AgentRunnerEnv` configured (P1 skipped).
- `CODEX_THREAD_ID=xyz789` set → Priority 2.
- `PI_CODING_AGENT=true` also set, but P2 fires first.

## Steps
1. Set `CODEX_THREAD_ID=xyz789` and `PI_CODING_AGENT=true`.
2. Priority 2 matches → returns `"codex"`, `true` (not "pi" from P3).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = []string{
        "CODEX_THREAD_ID=xyz789",
        "PI_CODING_AGENT=true",
    }
    return nil
}
```
