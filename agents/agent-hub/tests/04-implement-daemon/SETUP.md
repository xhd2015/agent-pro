## Preconditions
- The daemon package is available at `github.com/xhd2015/agent-pro/agents/agent-hub/daemon`.
- Each test uses a temporary Agent Hub home.

## Steps
1. Create a daemon home.
2. Execute the selected daemon lifecycle or queue scenario.
3. Verify socket, lock, storage, cursor, status, or restart behavior.

```go
import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"

    "github.com/xhd2015/agent-pro/agents/agent-hub/daemon"
    "github.com/xhd2015/agent-pro/agents/agent-hub/model"
)

type Request struct { Case string; Home string }
type Response struct { OK bool; Error string; Count int; Cursor model.Cursor; Status daemon.Status; SocketPath string; LockPath string }

func Setup(t *testing.T, req *Request) error {
    _ = expectOK
    _ = expectErrContains
    req.Case = "unset"
    req.Home = t.TempDir()
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    resp := &Response{}
    evt := model.NormalizedEvent{EventType:model.EventSessionStarted, Runner:"codex", RunnerSessionID:"s1"}
    switch req.Case {
    case "startup-creates-socket":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }; defer d.Stop()
        resp.SocketPath = d.SocketPath(); _, err := os.Stat(resp.SocketPath); resp.OK = err == nil
    case "startup-acquires-lock":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }; defer d.Stop()
        resp.LockPath = filepath.Join(req.Home, "daemon.lock"); _, err := os.Stat(resp.LockPath); resp.OK = err == nil
    case "startup-second-daemon-fails":
        d1 := daemon.New(req.Home); if err := d1.Start(); err != nil { t.Fatalf("start1: %v", err) }; defer d1.Stop()
        d2 := daemon.New(req.Home); err := d2.Start(); if err != nil { resp.Error = err.Error() }
    case "notify-appends-event":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }; defer d.Stop()
        env, err := d.Notify(evt); if err != nil { t.Fatalf("notify: %v", err) }
        resp.OK = env.Offset == 0 && env.Partition != ""
    case "notify-invalid-event-rejected":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }; defer d.Stop()
        _, err := d.Notify(model.NormalizedEvent{Runner:"codex"}); if err != nil { resp.Error = err.Error() }
    case "fetch-default-limit-one":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }; defer d.Stop()
        d.Notify(evt); d.Notify(model.NormalizedEvent{EventType:model.EventSessionFinished, Runner:"codex", RunnerSessionID:"s1"})
        out, err := d.Fetch("c1", 0, false); if err != nil { t.Fatalf("fetch: %v", err) }
        resp.Count = len(out.Events); resp.Cursor = out.NextCursor
    case "fetch-limit-ten":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }; defer d.Stop()
        for i:=0; i<3; i++ { d.Notify(model.NormalizedEvent{EventType:model.EventSessionStarted, Runner:"codex", RunnerSessionID:time.Now().String()}) }
        out, err := d.Fetch("c1", 10, false); if err != nil { t.Fatalf("fetch: %v", err) }
        resp.Count = len(out.Events); resp.Cursor = out.NextCursor
    case "fetch-peek-no-advance":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }; defer d.Stop()
        d.Notify(evt); out, err := d.Fetch("c1", 1, true); if err != nil { t.Fatalf("peek: %v", err) }
        out2, err := d.Fetch("c1", 1, false); if err != nil { t.Fatalf("fetch: %v", err) }
        resp.OK = len(out.Events)==1 && len(out2.Events)==1 && out.Events[0].Offset == out2.Events[0].Offset
    case "status-json-health":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }; defer d.Stop()
        st, err := d.Status(); if err != nil { t.Fatalf("status: %v", err) }
        resp.Status = st; resp.OK = st.Running && st.Home == req.Home
    case "restart-preserves-events":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }
        d.Notify(evt); d.Stop()
        d2 := daemon.New(req.Home); if err := d2.Start(); err != nil { t.Fatalf("restart: %v", err) }; defer d2.Stop()
        out, err := d2.Fetch("c1", 1, false); if err != nil { t.Fatalf("fetch: %v", err) }
        resp.Count = len(out.Events)
    case "restart-rebuilds-sessions":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }
        d.Notify(evt); os.RemoveAll(filepath.Join(req.Home, "sessions")); d.Stop()
        d2 := daemon.New(req.Home); if err := d2.Start(); err != nil { t.Fatalf("restart: %v", err) }; defer d2.Stop()
        _, err := os.Stat(filepath.Join(req.Home, "sessions", "active", "codex_s1.json")); resp.OK = err == nil
    case "shutdown-releases-lock":
        d := daemon.New(req.Home); if err := d.Start(); err != nil { t.Fatalf("start: %v", err) }
        if err := d.Stop(); err != nil { t.Fatalf("stop: %v", err) }
        _, err := os.Stat(filepath.Join(req.Home, "daemon.lock")); resp.OK = os.IsNotExist(err)
    default:
        resp.Error = "unknown case: "+req.Case
    }
    return resp, nil
}

func expectOK(t *testing.T, resp *Response, err error) { if err != nil || !resp.OK { t.Fatalf("resp=%+v err=%v", resp, err) } }
func expectErrContains(t *testing.T, resp *Response, want string) { if !strings.Contains(resp.Error, want) { t.Fatalf("error=%q want %q", resp.Error, want) } }
```
