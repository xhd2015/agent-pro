## Preconditions
- Production integration docs are expected at `agents/agent-hub/doc/RUNNER_INTEGRATION.md`.

## Steps
1. Read the production integration doc.
2. Check the selected required phrase or example.

```go
import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

type Request struct { Case string; DocPath string }
type Response struct { OK bool; Content string; Want string }

func Setup(t *testing.T, req *Request) error {
    req.Case = "unset"
    req.DocPath = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../doc/RUNNER_INTEGRATION.md"))
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    data, _ := os.ReadFile(req.DocPath)
    resp := &Response{Content:string(data)}
    wants := map[string]string{
        "codex-doc-session-start-example":"SessionStart",
        "codex-doc-user-prompt-example":"UserPromptSubmit",
        "codex-doc-stop-example":"Stop",
        "codex-doc-pre-tool-example":"PreToolUse",
        "codex-doc-post-tool-example":"PostToolUse",
        "codex-doc-trust-warning":"Codex hook trust",
        "opencode-doc-plugin-shape":"export const AgentHubPlugin",
        "opencode-doc-session-created":"session.created",
        "opencode-doc-session-error":"session.error",
        "opencode-doc-tool-events":"tool.execute.before",
        "daemon-doc-start-required":"agent-hub daemon start",
        "test-doc-fakes-only":"Automated tests use fake runners",
    }
    resp.Want = wants[req.Case]
    resp.OK = resp.Want != "" && strings.Contains(resp.Content, resp.Want)
    return resp, nil
}
```
