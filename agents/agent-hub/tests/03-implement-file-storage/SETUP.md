## Preconditions
- File storage is available at `github.com/xhd2015/agent-pro/agents/agent-hub/storage`.
- Each test uses a temporary Agent Hub home.

## Steps
1. Create a storage home in `t.TempDir`.
2. Execute the selected storage scenario.
3. Return observable offsets, cursors, paths, and session files.

```go
import (
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "github.com/xhd2015/agent-pro/agents/agent-hub/model"
    "github.com/xhd2015/agent-pro/agents/agent-hub/storage"
)

type Request struct {
    Case string
    Home string
}

type Response struct {
    OK bool
    Error string
    Offsets []int64
    Partitions []string
    Events []model.Envelope
    Cursor model.Cursor
    Path string
    Data string
}

func Setup(t *testing.T, req *Request) error {
    req.Case = "unset"
    req.Home = t.TempDir()
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    s := storage.New(req.Home)
    resp := &Response{}
    day1 := time.Date(2026, 6, 10, 1, 0, 0, 0, time.UTC)
    day2 := time.Date(2026, 6, 11, 1, 0, 0, 0, time.UTC)
    evt := func(t model.EventType, sid string) model.NormalizedEvent {
        return model.NormalizedEvent{EventType: t, Runner: "codex", RunnerSessionID: sid}
    }
    appendAt := func(ts time.Time, e model.NormalizedEvent) model.Envelope {
        env, err := s.Append(e, ts)
        if err != nil { t.Fatalf("append: %v", err) }
        return env
    }
    switch req.Case {
    case "append-first-event-offset-zero":
        env := appendAt(day1, evt(model.EventSessionStarted, "s1"))
        resp.Offsets = []int64{env.Offset}
        resp.Partitions = []string{env.Partition}
    case "append-second-event-offset-one":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        env := appendAt(day1, evt(model.EventSessionFinished, "s1"))
        resp.Offsets = []int64{env.Offset}
    case "append-new-day-resets-offset":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        env := appendAt(day2, evt(model.EventSessionStarted, "s2"))
        resp.Offsets = []int64{env.Offset}
        resp.Partitions = []string{env.Partition}
    case "append-received-at-selects-partition":
        env := appendAt(day2, model.NormalizedEvent{EventType: model.EventSessionStarted, Runner: "codex", RunnerSessionID: "s1", OccurredAt: day1})
        resp.Partitions = []string{env.Partition}
        resp.Path = filepath.Join(req.Home, "events", "2026", "06", "11", "events.jsonl")
        _, err := os.Stat(resp.Path)
        resp.OK = err == nil
    case "read-single-partition-batch":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        appendAt(day1, evt(model.EventSessionFinished, "s1"))
        events, cur, _, err := s.ReadBatch(model.Cursor{Partition:"2026-06-10", Offset:0}, 2)
        if err != nil { t.Fatalf("read: %v", err) }
        resp.Events = events; resp.Cursor = cur
    case "read-cross-partition-batch":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        appendAt(day2, evt(model.EventSessionStarted, "s2"))
        events, cur, _, err := s.ReadBatch(model.Cursor{Partition:"2026-06-10", Offset:0}, 2)
        if err != nil { t.Fatalf("read: %v", err) }
        resp.Events = events; resp.Cursor = cur
    case "read-missing-day-skipped":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        appendAt(day2.AddDate(0,0,2), evt(model.EventSessionStarted, "s3"))
        events, cur, _, err := s.ReadBatch(model.Cursor{Partition:"2026-06-11", Offset:0}, 1)
        if err != nil { t.Fatalf("read: %v", err) }
        resp.Events = events; resp.Cursor = cur
    case "cursor-save-load":
        want := model.Cursor{Partition:"2026-06-10", Offset:7}
        if err := s.SaveCursor("c1", want); err != nil { t.Fatalf("save cursor: %v", err) }
        got, err := s.LoadCursor("c1")
        if err != nil { t.Fatalf("load cursor: %v", err) }
        resp.Cursor = got; resp.OK = got == want
    case "cursor-advance-after-batch":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        appendAt(day1, evt(model.EventSessionFinished, "s1"))
        events, err := s.Fetch("c1", 2, false)
        if err != nil { t.Fatalf("fetch: %v", err) }
        cur, err := s.LoadCursor("c1")
        if err != nil { t.Fatalf("load cursor: %v", err) }
        resp.Events = events.Events; resp.Cursor = cur
    case "index-rebuild-missing":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        idx := filepath.Join(req.Home, "events", "2026", "06", "10", "events.idx")
        if err := os.Remove(idx); err != nil { t.Fatalf("remove idx: %v", err) }
        if err := s.RebuildIndexes(); err != nil { t.Fatalf("rebuild: %v", err) }
        data, err := os.ReadFile(idx)
        if err != nil { t.Fatalf("read idx: %v", err) }
        resp.Data = string(data); resp.OK = strings.Contains(resp.Data, `"offset":0`)
    case "session-project-started":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        p := filepath.Join(req.Home, "sessions", "active", "codex_s1.json")
        data, err := os.ReadFile(p)
        if err != nil { t.Fatalf("read session: %v", err) }
        resp.Path = p; resp.Data = string(data); resp.OK = strings.Contains(resp.Data, `"status":"active"`)
    case "session-project-finished":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        appendAt(day2, evt(model.EventSessionFinished, "s1"))
        p := filepath.Join(req.Home, "sessions", "completed", "codex_s1.json")
        data, err := os.ReadFile(p)
        if err != nil { t.Fatalf("read session: %v", err) }
        resp.Path = p; resp.Data = string(data); resp.OK = strings.Contains(resp.Data, `"status":"completed"`)
    case "recovery-rebuild-sessions":
        appendAt(day1, evt(model.EventSessionStarted, "s1"))
        if err := os.RemoveAll(filepath.Join(req.Home, "sessions")); err != nil { t.Fatalf("remove sessions: %v", err) }
        if err := s.RebuildSessions(); err != nil { t.Fatalf("rebuild sessions: %v", err) }
        data, err := os.ReadFile(filepath.Join(req.Home, "sessions", "active", "codex_s1.json"))
        if err != nil { t.Fatalf("read session: %v", err) }
        var obj map[string]any
        if err := json.Unmarshal(data, &obj); err != nil { t.Fatalf("session json: %v", err) }
        resp.OK = obj["status"] == "active"
    default:
        resp.Error = "unknown case: " + req.Case
    }
    return resp, nil
}
```

