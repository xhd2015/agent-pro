## Expected
- When running under opencode: exit 0, meta.json has `main_agent_opencode_session_id` with `ses_*` value.
- When NOT running under opencode: discovery fails, test skips.

```go
import (
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if resp.ExitCode != 0 {
        if strings.Contains(resp.Stderr, "must be run from codex or opencode") {
            t.Skip("not running under opencode; auto-discovery requires opencode ancestor in process tree")
        }
        t.Fatalf("exit code = %d, want 0\nstderr:\n%s", resp.ExitCode, resp.Stderr)
    }

    sessionsDir := sessionsDir()

    today := time.Now().Format("2006/01/02")
    dateDir := filepath.Join(sessionsDir, today)
    entries, readErr := os.ReadDir(dateDir)
    if readErr != nil {
        t.Fatalf("cannot read date dir %s: %v", dateDir, readErr)
    }

    for _, entry := range entries {
        if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "sess_") {
            continue
        }
        metaPath := filepath.Join(dateDir, entry.Name(), "meta.json")
        data, readErr := os.ReadFile(metaPath)
        if readErr != nil {
            continue
        }
        var meta struct {
            MainAgentOpencodeSessionID string `json:"main_agent_opencode_session_id"`
        }
        if jsonErr := json.Unmarshal(data, &meta); jsonErr != nil {
            continue
        }
        if meta.MainAgentOpencodeSessionID != "" {
            if !strings.HasPrefix(meta.MainAgentOpencodeSessionID, "ses_") {
                t.Fatalf("main_agent_opencode_session_id must start with ses_, got %q", meta.MainAgentOpencodeSessionID)
            }
            t.Logf("discovered session: %s", meta.MainAgentOpencodeSessionID)
            return
        }
    }
    t.Fatal("no session found with main_agent_opencode_session_id set")
}
```
