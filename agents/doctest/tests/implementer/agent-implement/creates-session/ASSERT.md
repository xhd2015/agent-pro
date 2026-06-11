## Expected
- Exit code 0.
- Session directory is created under `~/.agent-pro/dedicated-agents/doctest-agent-implementer/sessions/YYYY/MM/DD/sess_*/`.
- The session dir contains `meta.json` with `codex_thread_id`.
- A `questions/` dir exists with a timestamp-named `.json` file (created before the agent runs).

## Exit Code
- Exit code 0.

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
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code = %d\nstderr:\n%s", resp.ExitCode, resp.Stderr)
    }

    home, _ := os.UserHomeDir()
    sessionsDir := filepath.Join(home, ".agent-pro", "dedicated-agents", "doctest-agent-implementer", "sessions")

    today := time.Now().Format("2006/01/02")
    dateDir := filepath.Join(sessionsDir, today)
    entries, readErr := os.ReadDir(dateDir)
    if readErr != nil {
        t.Fatalf("cannot read date dir %s: %v", dateDir, readErr)
    }

    var newestDir string
    var newestTime int64
    for _, entry := range entries {
        if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "sess_") {
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

    sessDir := filepath.Join(dateDir, newestDir)

    metaPath := filepath.Join(sessDir, "meta.json")
    data, readErr := os.ReadFile(metaPath)
    if readErr != nil {
        t.Fatalf("cannot read meta.json: %v", readErr)
    }
    var meta struct {
        CodexThreadID string `json:"codex_thread_id"`
        CreatedAt     string `json:"created_at"`
    }
    if jsonErr := json.Unmarshal(data, &meta); jsonErr != nil {
        t.Fatalf("invalid meta.json: %v", jsonErr)
    }
    if meta.CodexThreadID == "" {
        t.Fatal("meta.json missing codex_thread_id")
    }
    if meta.CreatedAt == "" {
        t.Fatal("meta.json missing created_at")
    }

    questionsDir := filepath.Join(sessDir, "questions")
    qFiles, qErr := os.ReadDir(questionsDir)
    if qErr != nil {
        t.Fatalf("cannot read questions dir %s: %v", questionsDir, qErr)
    }
    var jsonFiles []string
    for _, f := range qFiles {
        if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
            jsonFiles = append(jsonFiles, f.Name())
        }
    }
    if len(jsonFiles) == 0 {
        t.Fatal("no questions .json file found")
    }
    match := false
    for _, name := range jsonFiles {
        if matched, _ := filepath.Match("[0-9][0-9][0-9][0-9]_[0-9][0-9]_[0-9][0-9]_[0-9][0-9]_[0-9][0-9]_[0-9][0-9]*.json", name); matched {
            match = true
            break
        }
    }
    if !match {
        t.Fatalf("no timestamp-named .json file found in questions dir, got: %v", jsonFiles)
    }
}
```
