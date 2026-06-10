## Preconditions
- `cmd/agent-hub` exists and supports `hook notify`.
- Each test uses a temporary Agent Hub home.

## Steps
1. Build `cmd/agent-hub`.
2. Send one native hook payload to `agent-hub hook notify`.
3. Fetch the normalized event or assert the expected error.

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

type Request struct { Case string; Home string; RepoRoot string; Bin string }
type Response struct { OK bool; EventType string; Runner string; Prompt string; Error string; ExitCode int; Stderr string }

func Setup(t *testing.T, req *Request) error {
    _ = runHub
    _ = hookAndFetch
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
    switch req.Case {
    case "codex-session-start": return hookAndFetch(t, req, "codex", "SessionStart", `{"session_id":"c1"}`), nil
    case "codex-user-prompt-submit": return hookAndFetch(t, req, "codex", "UserPromptSubmit", `{"session_id":"c1","prompt":"hello codex"}`), nil
    case "codex-stop": return hookAndFetch(t, req, "codex", "Stop", `{"session_id":"c1"}`), nil
    case "codex-pre-tool-use": return hookAndFetch(t, req, "codex", "PreToolUse", `{"session_id":"c1","tool":"Bash"}`), nil
    case "codex-post-tool-use": return hookAndFetch(t, req, "codex", "PostToolUse", `{"session_id":"c1","tool":"Bash"}`), nil
    case "codex-permission-request": return hookAndFetch(t, req, "codex", "PermissionRequest", `{"session_id":"c1","tool":"Bash"}`), nil
    case "codex-subagent-start": return hookAndFetch(t, req, "codex", "SubagentStart", `{"session_id":"c1"}`), nil
    case "opencode-session-created": return hookAndFetch(t, req, "opencode", "session.created", `{"sessionID":"o1"}`), nil
    case "opencode-message-updated": return hookAndFetch(t, req, "opencode", "message.updated", `{"sessionID":"o1","message":{"role":"user","text":"hello opencode"}}`), nil
    case "opencode-session-idle": return hookAndFetch(t, req, "opencode", "session.idle", `{"sessionID":"o1"}`), nil
    case "opencode-session-error": return hookAndFetch(t, req, "opencode", "session.error", `{"sessionID":"o1","error":"bad"}`), nil
    case "opencode-tool-before": return hookAndFetch(t, req, "opencode", "tool.execute.before", `{"sessionID":"o1","tool":"bash"}`), nil
    case "opencode-tool-after": return hookAndFetch(t, req, "opencode", "tool.execute.after", `{"sessionID":"o1","tool":"bash"}`), nil
    case "errors-unknown-runner":
        out := runHub(t, req, `{"session_id":"x"}`, "hook", "notify", "--runner", "bad", "--event", "SessionStart")
        resp.ExitCode = out.ExitCode; resp.Stderr = out.Stderr; resp.OK = out.ExitCode != 0 && strings.Contains(out.Stderr, "unknown runner")
    case "errors-unknown-event":
        out := runHub(t, req, `{"session_id":"x"}`, "hook", "notify", "--runner", "codex", "--event", "Nope")
        resp.ExitCode = out.ExitCode; resp.Stderr = out.Stderr; resp.OK = out.ExitCode != 0 && strings.Contains(out.Stderr, "unknown hook event")
    default:
        resp.Error = "unknown case: "+req.Case
    }
    return resp, nil
}

type cliOut struct { ExitCode int; Stdout string; Stderr string }

func runHub(t *testing.T, req *Request, stdin string, args ...string) cliOut {
    t.Helper()
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second); defer cancel()
    cmd := exec.CommandContext(ctx, req.Bin, args...)
    cmd.Dir = req.Home
    cmd.Env = append(os.Environ(), "AGENT_HUB_HOME="+req.Home)
    if stdin != "" { cmd.Stdin = strings.NewReader(stdin) }
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout; cmd.Stderr = &stderr
    err := cmd.Run()
    out := cliOut{Stdout:stdout.String(), Stderr:stderr.String()}
    if err != nil { var ee *exec.ExitError; if errors.As(err, &ee) { out.ExitCode = ee.ExitCode() } else { out.ExitCode = 1 } }
    return out
}

func hookAndFetch(t *testing.T, req *Request, runner string, event string, payload string) *Response {
    t.Helper()
    out := runHub(t, req, payload, "hook", "notify", "--runner", runner, "--event", event)
    if out.ExitCode != 0 { return &Response{ExitCode:out.ExitCode, Stderr:out.Stderr, Error:out.Stderr} }
    fetched := runHub(t, req, "", "fetch", "--consumer-id", "test")
    if fetched.ExitCode != 0 { return &Response{ExitCode:fetched.ExitCode, Stderr:fetched.Stderr, Error:fetched.Stderr} }
    var obj map[string]any
    if err := json.Unmarshal([]byte(fetched.Stdout), &obj); err != nil { t.Fatalf("fetch json: %v\n%s", err, fetched.Stdout) }
    events := obj["events"].([]any)
    env := events[0].(map[string]any)
    e := env["event"].(map[string]any)
    prompt, _ := e["prompt"].(string)
    return &Response{OK:true, EventType:e["event_type"].(string), Runner:e["runner"].(string), Prompt:prompt}
}
```

