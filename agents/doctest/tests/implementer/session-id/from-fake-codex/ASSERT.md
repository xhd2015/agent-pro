## Expected
- Exit code 0.
- Meta.json has `main_agent_codex_thread_id` set to `codex-tid-222`.
- Meta.json has `doctest_agent_implementer_session_id` set to `codex-tid-222` (fallback routing key).
- Meta.json has `opencode_session_id` set.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d, want 0\nstderr:\n%s", resp.ExitCode, resp.Stderr)
    }

    metaPath := findSessionMeta(t, "main_agent_codex_thread_id", "codex-tid-222")
    if metaPath == "" {
        t.Fatal("no session found with main_agent_codex_thread_id=codex-tid-222")
    }

    m := readMetaJSON(t, metaPath)

    if v, ok := m["main_agent_codex_thread_id"].(string); !ok || v != "codex-tid-222" {
        t.Fatalf("main_agent_codex_thread_id = %v, want codex-tid-222", m["main_agent_codex_thread_id"])
    }
    if v, ok := m["doctest_agent_implementer_session_id"].(string); !ok || v != "codex-tid-222" {
        t.Fatalf("doctest_agent_implementer_session_id = %v, want codex-tid-222", m["doctest_agent_implementer_session_id"])
    }
    if v, ok := m["main_agent_opencode_session_id"]; ok {
        t.Fatalf("main_agent_opencode_session_id should NOT be present when using codex, got %v", v)
    }
    if v, ok := m["opencode_session_id"].(string); !ok || v == "" {
        t.Fatalf("opencode_session_id must be set, got %v", m["opencode_session_id"])
    }
    if _, ok := m["created_at"]; !ok {
        t.Fatal("created_at must be set")
    }

    t.Logf("stdout: %s", strings.TrimSpace(resp.Stdout))
}
```
