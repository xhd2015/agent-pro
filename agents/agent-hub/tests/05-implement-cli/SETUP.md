## Preconditions
- The repository contains `cmd/agent-hub`.
- Each test uses a temporary `AGENT_HUB_HOME`.

## Steps
1. Build `cmd/agent-hub`.
2. Execute the selected CLI scenario.
3. Capture stdout, stderr, exit status, and stored queue state.

```go
import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
    "time"
)

type Request struct { Case string; RepoRoot string; Home string; Bin string }
type Response struct { OK bool; ExitCode int; Stdout string; Stderr string; Error string; JSON map[string]any }

func Setup(t *testing.T, req *Request) error {
    _ = runHubWithStdin
    _ = runHub
    _ = mustRun
    _ = parseJSON
    req.Case = "unset"
    req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
    req.Home = t.TempDir()
    req.Bin = filepath.Join(req.Home, "agent-hub")
    build := exec.Command("go", "build", "-o", req.Bin, "./cmd/agent-hub")
    build.Dir = req.RepoRoot
    if out, err := build.CombinedOutput(); err != nil { return fmt.Errorf("build agent-hub: %w\n%s", err, string(out)) }
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    resp := &Response{}
    valid := `{"event_type":"agent.session.started","runner":"codex","runner_session_id":"s1"}`
    switch req.Case {
    case "daemon-start-status-stop":
        mustRun(t, req, "daemon", "start")
        st := mustRun(t, req, "daemon", "status", "--json")
        mustRun(t, req, "daemon", "stop")
        resp.Stdout = st.Stdout; resp.JSON = parseJSON(t, st.Stdout); resp.OK = resp.JSON["running"] == true
    case "daemon-status-json":
        mustRun(t, req, "daemon", "start")
        out := mustRun(t, req, "daemon", "status", "--json")
        resp.JSON = parseJSON(t, out.Stdout); resp.OK = resp.JSON["home"] == req.Home
    case "notify-json-valid":
        out := mustRun(t, req, "notify", "--json", valid)
        resp.JSON = parseJSON(t, out.Stdout); resp.OK = resp.JSON["offset"] == float64(0)
    case "notify-file-valid":
        p := filepath.Join(req.Home, "event.json"); os.WriteFile(p, []byte(valid), 0644)
        out := mustRun(t, req, "notify", "--file", p)
        resp.JSON = parseJSON(t, out.Stdout); resp.OK = resp.JSON["offset"] == float64(0)
    case "notify-json-invalid":
        out := runHub(t, req, "notify", "--json", "{")
        resp.ExitCode = out.ExitCode; resp.Stderr = out.Stderr; resp.OK = out.ExitCode != 0 && strings.Contains(out.Stderr, "invalid")
    case "hook-notify-codex-session-start":
        out := runHubWithStdin(t, req, `{"session_id":"s1"}`, "hook", "notify", "--runner", "codex", "--event", "SessionStart")
        resp.Stdout = out.Stdout; resp.Stderr = out.Stderr; resp.ExitCode = out.ExitCode; resp.OK = out.ExitCode == 0
    case "hook-notify-opencode-session-created":
        out := runHubWithStdin(t, req, `{"sessionID":"s1"}`, "hook", "notify", "--runner", "opencode", "--event", "session.created")
        resp.Stdout = out.Stdout; resp.Stderr = out.Stderr; resp.ExitCode = out.ExitCode; resp.OK = out.ExitCode == 0
    case "fetch-default-limit":
        mustRun(t, req, "notify", "--json", valid); mustRun(t, req, "notify", "--json", valid)
        out := mustRun(t, req, "fetch", "--consumer-id", "c1")
        resp.JSON = parseJSON(t, out.Stdout); resp.OK = len(resp.JSON["events"].([]any)) == 1
    case "fetch-limit-ten":
        for i:=0; i<3; i++ { mustRun(t, req, "notify", "--json", valid) }
        out := mustRun(t, req, "fetch", "--consumer-id", "c1", "--limit", "10")
        resp.JSON = parseJSON(t, out.Stdout); resp.OK = len(resp.JSON["events"].([]any)) == 3
    case "fetch-limit-zero-invalid":
        out := runHub(t, req, "fetch", "--consumer-id", "c1", "--limit", "0")
        resp.ExitCode = out.ExitCode; resp.Stderr = out.Stderr; resp.OK = out.ExitCode != 0 && strings.Contains(out.Stderr, "limit")
    case "fetch-peek":
        mustRun(t, req, "notify", "--json", valid)
        a := mustRun(t, req, "fetch", "--consumer-id", "c1", "--peek")
        b := mustRun(t, req, "fetch", "--consumer-id", "c1")
        resp.OK = strings.Contains(a.Stdout, `"offset":0`) && strings.Contains(b.Stdout, `"offset":0`)
    case "replay-from-cursor":
        mustRun(t, req, "notify", "--json", valid); mustRun(t, req, "fetch", "--consumer-id", "c1")
        out := mustRun(t, req, "replay", "--consumer-id", "c1", "--from", time.Now().UTC().Format("2006-01-02")+":0")
        resp.JSON = parseJSON(t, out.Stdout); resp.OK = resp.JSON["consumer_id"] == "c1"
    case "inspection-consumers":
        mustRun(t, req, "notify", "--json", valid); mustRun(t, req, "fetch", "--consumer-id", "c1")
        out := mustRun(t, req, "consumers", "--json")
        resp.Stdout = out.Stdout; resp.OK = strings.Contains(out.Stdout, "c1")
    case "inspection-sessions":
        mustRun(t, req, "notify", "--json", valid)
        out := mustRun(t, req, "sessions", "--json")
        resp.Stdout = out.Stdout; resp.OK = strings.Contains(out.Stdout, "s1")
    default:
        resp.Error = "unknown case: "+req.Case
    }
    return resp, nil
}

type cliOut struct { ExitCode int; Stdout string; Stderr string; Err error }

func runHubWithStdin(t *testing.T, req *Request, stdin string, args ...string) cliOut {
    t.Helper()
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second); defer cancel()
    cmd := exec.CommandContext(ctx, req.Bin, args...)
    cmd.Dir = req.Home
    cmd.Env = append(os.Environ(), "AGENT_HUB_HOME="+req.Home)
    if stdin != "" { cmd.Stdin = strings.NewReader(stdin) }
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout; cmd.Stderr = &stderr
    err := cmd.Run()
    out := cliOut{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
    if err != nil { var ee *exec.ExitError; if errors.As(err, &ee) { out.ExitCode = ee.ExitCode() } }
    return out
}

func runHub(t *testing.T, req *Request, args ...string) cliOut { return runHubWithStdin(t, req, "", args...) }

func mustRun(t *testing.T, req *Request, args ...string) cliOut {
    t.Helper()
    out := runHub(t, req, args...)
    if out.ExitCode != 0 || out.Err != nil { t.Fatalf("agent-hub %v failed: code=%d err=%v stderr=%s", args, out.ExitCode, out.Err, out.Stderr) }
    return out
}

func parseJSON(t *testing.T, text string) map[string]any {
    t.Helper()
    var obj map[string]any
    if err := json.Unmarshal([]byte(text), &obj); err != nil { t.Fatalf("parse json: %v\n%s", err, text) }
    return obj
}
```
