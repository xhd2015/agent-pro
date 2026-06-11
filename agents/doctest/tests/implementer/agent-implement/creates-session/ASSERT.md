## Expected
- Exit code 0.
- Session directory is created under `~/.agent-pro/dedicated-agents/doctest-agent/sessions/`.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d\nstderr:\n%s", resp.ExitCode, resp.Stderr)
    }

    home, _ := os.UserHomeDir()
    agentDir := filepath.Join(home, ".agent-pro", "dedicated-agents", "doctest-agent", "sessions")
    entries, readErr := os.ReadDir(agentDir)
    if readErr != nil {
        t.Fatalf("cannot read session dir %s: %v", agentDir, readErr)
    }
    var newestDir string
    var newestTime int64
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }
        info, statErr := entry.Info()
        if statErr != nil {
            continue
        }
        modTime := info.ModTime().UnixNano()
        if modTime > newestTime {
            newestTime = modTime
            newestDir = entry.Name()
        }
    }
    if newestDir == "" {
        t.Fatal("no session directory created")
    }
}
```
