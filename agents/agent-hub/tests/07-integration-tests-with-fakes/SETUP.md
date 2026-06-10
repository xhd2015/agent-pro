## Preconditions
- `cmd/agent-hub`, `cmd/fake-codex`, and `cmd/fake-opencode` build locally.
- Tests use only temporary config and storage directories.

## Steps
1. Build the three binaries into a temporary directory.
2. Run fake runners with `--mock-config`.
3. Fetch events from Agent Hub.
4. Verify event types, cursor behavior, partition files, and isolation.

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

type Request struct { Case string; RepoRoot string; TempDir string; Home string; Hub string; FakeCodex string; FakeOpencode string }
type Response struct { OK bool; Error string; Events []string; Count int; Stdout string }

func Setup(t *testing.T, req *Request) error {
    _ = buildBin; _ = runCmd; _ = writeFile; _ = fetchTypes; _ = runFakeCodex; _ = runFakeOpencode
    req.Case = "unset"
    req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
    req.TempDir = t.TempDir()
    req.Home = filepath.Join(req.TempDir, "hub")
    req.Hub = filepath.Join(req.TempDir, "agent-hub")
    req.FakeCodex = filepath.Join(req.TempDir, "fake-codex")
    req.FakeOpencode = filepath.Join(req.TempDir, "fake-opencode")
    buildBin(t, req, req.Hub, "./cmd/agent-hub")
    buildBin(t, req, req.FakeCodex, "./cmd/fake-codex")
    buildBin(t, req, req.FakeOpencode, "./cmd/fake-opencode")
    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    resp := &Response{}
    switch req.Case {
    case "fake-codex-lifecycle-success":
        runFakeCodex(t, req, codexMock(req, 0, []string{"SessionStart","UserPromptSubmit","Stop"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = hasTypes(resp.Events, "agent.session.started", "agent.prompt.submitted", "agent.session.finished")
    case "fake-codex-lifecycle-failure":
        runFakeCodex(t, req, codexMock(req, 7, []string{"SessionStart","Error"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = hasTypes(resp.Events, "agent.session.started", "agent.session.failed")
    case "fake-codex-tool-events":
        runFakeCodex(t, req, codexMock(req, 0, []string{"PreToolUse","PostToolUse"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = hasTypes(resp.Events, "agent.tool.started", "agent.tool.finished")
    case "fake-codex-batch-fetch":
        runFakeCodex(t, req, codexMock(req, 0, []string{"SessionStart","UserPromptSubmit","Stop"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = len(resp.Events)==3
    case "fake-opencode-lifecycle-success":
        runFakeOpencode(t, req, opencodeMock(req, 0, []string{"session.created","message.updated","session.idle"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = hasTypes(resp.Events, "agent.session.started", "agent.prompt.submitted", "agent.session.finished")
    case "fake-opencode-lifecycle-failure":
        runFakeOpencode(t, req, opencodeMock(req, 6, []string{"session.created","session.error"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = hasTypes(resp.Events, "agent.session.started", "agent.session.failed")
    case "fake-opencode-tool-events":
        runFakeOpencode(t, req, opencodeMock(req, 0, []string{"tool.execute.before","tool.execute.after"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = hasTypes(resp.Events, "agent.tool.started", "agent.tool.finished")
    case "fake-opencode-batch-fetch":
        runFakeOpencode(t, req, opencodeMock(req, 0, []string{"session.created","message.updated","session.idle"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = len(resp.Events)==3
    case "mixed-runners-same-daemon":
        runFakeCodex(t, req, codexMock(req, 0, []string{"SessionStart"})); runFakeOpencode(t, req, opencodeMock(req, 0, []string{"session.created"}))
        resp.Events = fetchTypes(t, req, "c1", "10"); resp.OK = len(resp.Events)==2
    case "storage-date-partition-created":
        runFakeCodex(t, req, codexMock(req, 0, []string{"SessionStart"}))
        matches, _ := filepath.Glob(filepath.Join(req.Home, "events", "*", "*", "*", "events.jsonl")); resp.OK = len(matches)==1
    case "cursor-independent-consumers":
        runFakeCodex(t, req, codexMock(req, 0, []string{"SessionStart"}))
        a := fetchTypes(t, req, "a", "1"); b := fetchTypes(t, req, "b", "1"); resp.OK = len(a)==1 && len(b)==1
    case "isolation-no-host-config-mutation":
        runFakeCodex(t, req, codexMock(req, 0, []string{"SessionStart"})); runFakeOpencode(t, req, opencodeMock(req, 0, []string{"session.created"}))
        _, codexErr := os.Stat(filepath.Join(req.TempDir, "home", ".codex")); _, opErr := os.Stat(filepath.Join(req.TempDir, "home", ".config", "opencode")); resp.OK = os.IsNotExist(codexErr) && os.IsNotExist(opErr)
    default:
        resp.Error = "unknown case: "+req.Case
    }
    return resp, nil
}

func buildBin(t *testing.T, req *Request, out string, pkg string) { t.Helper(); cmd := exec.Command("go","build","-o",out,pkg); cmd.Dir=req.RepoRoot; if b,err:=cmd.CombinedOutput(); err!=nil { t.Fatalf("build %s: %v\n%s", pkg, err, string(b)) } }

func runCmd(t *testing.T, req *Request, bin string, stdin string, args ...string) (string, int) {
    t.Helper(); ctx,cancel:=context.WithTimeout(context.Background(),15*time.Second); defer cancel()
    cmd:=exec.CommandContext(ctx, bin, args...); cmd.Dir=req.TempDir
    cmd.Env=append(os.Environ(),"AGENT_HUB_HOME="+req.Home,"HOME="+filepath.Join(req.TempDir,"home"),"CODEX_HOME="+filepath.Join(req.TempDir,"codex-home"),"OPENCODE_CONFIG_DIR="+filepath.Join(req.TempDir,"opencode-config"))
    if stdin!="" { cmd.Stdin=strings.NewReader(stdin) }
    var stdout,stderr bytes.Buffer; cmd.Stdout=&stdout; cmd.Stderr=&stderr
    err:=cmd.Run(); code:=0; if err!=nil { var ee *exec.ExitError; if errors.As(err,&ee){code=ee.ExitCode()} else {t.Fatalf("run %s: %v", bin, err)} }
    _=stderr.String(); return stdout.String(), code
}

func writeFile(t *testing.T, path string, content string) { t.Helper(); os.MkdirAll(filepath.Dir(path),0755); if err:=os.WriteFile(path,[]byte(content),0644); err!=nil { t.Fatal(err) } }

func runFakeCodex(t *testing.T, req *Request, mock string) { p:=filepath.Join(req.TempDir,"codex-mock.json"); writeFile(t,p,mock); _,_ = runCmd(t,req,req.FakeCodex,"","exec","--json","--mock-config",p,"prompt") }
func runFakeOpencode(t *testing.T, req *Request, mock string) { p:=filepath.Join(req.TempDir,"opencode-mock.json"); writeFile(t,p,mock); _,_ = runCmd(t,req,req.FakeOpencode,"","run","--format","json","--mock-config",p,"prompt") }

func fetchTypes(t *testing.T, req *Request, consumer string, limit string) []string {
    out,_:=runCmd(t,req,req.Hub,"","fetch","--consumer-id",consumer,"--limit",limit)
    var obj map[string]any; if err:=json.Unmarshal([]byte(out),&obj); err!=nil { t.Fatalf("fetch json: %v\n%s",err,out) }
    var types []string; for _,raw:= range obj["events"].([]any) { env:=raw.(map[string]any); event:=env["event"].(map[string]any); types=append(types,event["event_type"].(string)) }
    return types
}

func hasTypes(got []string, want ...string) bool { if len(got)<len(want){return false}; for i,w:=range want{ if got[i]!=w{return false} }; return true }

func hookJSON(hub string, runner string, events []string) string {
    var hooks []map[string]any
    for _, e := range events {
        at := "before_stdout"; if e=="Stop" || e=="session.idle" { at="after_stdout" }; if e=="Error" || e=="session.error" { at="on_error" }
        payload := map[string]any{"session_id":"s1","sessionID":"s1","prompt":"prompt text","message":map[string]any{"text":"prompt text"}}
        hooks = append(hooks, map[string]any{"at":at,"event":e,"payload":payload})
    }
    data,_:=json.Marshal(hooks); return string(data)
}

func codexMock(req *Request, exit int, events []string) string {
    hooks := hookJSON(req.Hub, "fake-codex", events)
    return fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","session_id":"fc1","exit_code":%d,"ignore_hook_errors":false,"hook_command":%q,"hooks":%s,"stdout_events":[]}`, exit, req.Hub+" hook notify --runner fake-codex --event {{event}}", hooks)
}

func opencodeMock(req *Request, exit int, events []string) string {
    hooks := hookJSON(req.Hub, "fake-opencode", events)
    return fmt.Sprintf(`{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"fo1","exit_code":%d,"ignore_hook_errors":false,"hook_command":%q,"hooks":%s,"stdout_events":[]}`, exit, req.Hub+" hook notify --runner fake-opencode --event {{event}}", hooks)
}
```
