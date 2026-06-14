## Preconditions
- `bun` is installed on the system (test fails if not found).
- `OPENCODE_CONFIG_DIR` and `AGENT_HUB_HOME` point to isolated temp directories.
- `AGENT_HUB_OPENCODE_RUNNER=fake-opencode` is set so hook events are stored under the fake-opencode runner.

## Steps
1. Install the agent-hub plugin via `--opencode-home` into fake-opencode's config home.
2. Overwrite the installed plugin with a test version that forwards `session.created`, `session.idle`, and `session.error` via `agent-hub hook notify`.
3. Write mock config run1: [step_start, sleep(3s), message "working on it", done].
4. Write mock config run2 (resume): [message "resumed work"].
5. The Run function starts run1 in the background, performs mid-flight checks (status=running, first events fetchable), waits for completion, and returns the response.

```go
import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
)

func Setup(t *testing.T, req *Request) error {
    if _, err := exec.LookPath("bun"); err != nil {
        t.Fatalf("full-workflow test requires bun: %v", err)
    }

    req.Env = append(req.Env,
        "AGENT_HUB_OPENCODE_RUNNER=fake-opencode",
    )

    opencodeHome := filepath.Join(req.TempDir, "opencode-home")
    req.Env = append(req.Env, "OPENCODE_CONFIG_DIR="+opencodeHome)

    // install plugin via --opencode-home
    instResp, err := runAgentHub(t, req, "integration", "opencode", "install", "--opencode-home", opencodeHome)
    if err != nil {
        return fmt.Errorf("install plugin: %w", err)
    }
    if instResp.ExitCode != 0 {
        return fmt.Errorf("install plugin failed (exit %d): %s", instResp.ExitCode, instResp.Stderr)
    }

    // overwrite with test plugin that forwards session.created and session.idle
    pluginPath := filepath.Join(opencodeHome, "plugins", "agent-hub.ts")
    testPlugin := `const { execSync } = require("child_process");
function notify(eventType, payload) {
  try {
    execSync("agent-hub hook notify --runner opencode --event " + eventType, { input: JSON.stringify(payload), stdio: ["pipe","pipe","pipe"] });
  } catch(e) {}
}
export const AgentHubPlugin = async () => {
  return {
    "session.created": async (event) => {
      notify("session.created", event);
    },
    "session.idle": async (event) => {
      notify("session.idle", event);
    },
    "session.error": async (event) => {
      notify("session.error", event);
    },
  };
};
`
    writeFile(t, pluginPath, testPlugin)

    // write run1 mock config with sleep for mid-flight checks
    writeFile(t, filepath.Join(req.TempDir, "run1-mock.json"), `{"version":"agent-pro.fake-runner.v1","runner":"opencode","session_id":"sess_full","llm_events":[{"type":"step_start"},{"type":"sleep","delay_ms":3000},{"type":"message","text":"working on it"},{"type":"done"}]}`)

    // write run2 mock config (resume)
    writeFile(t, filepath.Join(req.TempDir, "run2-mock.json"), `{"version":"agent-pro.fake-runner.v1","runner":"opencode","session_id":"sess_full","llm_events":[{"type":"message","text":"resumed work"}]}`)

    req.Command = req.FakeOpencode
    req.Args = []string{"run", "--format", "json", "--mock-config", filepath.Join(req.TempDir, "run1-mock.json"), "--plugin", pluginPath, "do the task"}

    return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, req.Command, req.Args...)
    cmd.Dir = req.TempDir
    cmd.Env = append(os.Environ(), req.Env...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    mkErrResp := func(format string, args ...any) (*Response, error) {
        msg := fmt.Sprintf(format, args...)
        return &Response{Stdout: stdout.String(), Stderr: stderr.String() + msg, Err: fmt.Errorf("%s", msg)}, fmt.Errorf("%s", msg)
    }

    if err := cmd.Start(); err != nil {
        return mkErrResp("cmd.Start: %v", err)
    }

    // wait for first events (step_start + session.created hook) to be emitted
    time.Sleep(1000 * time.Millisecond)

    // mid-flight check: session status must be running
    sr, err := runAgentHub(t, req, "session", "show", "--runner", "fake-opencode", "--session-id", "sess_full")
    if err != nil {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("session show (mid-flight): %v", err)
    }
    if sr.ExitCode != 0 {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("session show failed (exit %d): %s", sr.ExitCode, sr.Stderr)
    }
    so := parseJSON(t, sr.Stdout)
    if so["status"] != "running" {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("expected session status running mid-flight, got %v", so["status"])
    }

    // mid-flight fetch: verify session.started event exists
    fr, err := runAgentHub(t, req, "fetch", "--consumer-id", "consumer-full-"+t.Name(), "--limit", "10")
    if err != nil {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("fetch (mid-flight): %v", err)
    }
    if fr.ExitCode != 0 {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("fetch failed (exit %d): %s", fr.ExitCode, fr.Stderr)
    }
    frObj := parseJSON(t, fr.Stdout)
    events, _ := frObj["events"].([]any)
    if events == nil || len(events) == 0 {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("expected at least 1 event mid-flight, got 0")
    }
    hasStarted := false
    for _, e := range events {
        ev := e.(map[string]any)
        nev, _ := ev["event"].(map[string]any)
        if nev["event_type"] == "agent.session.started" {
            hasStarted = true
        }
    }
    if !hasStarted {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
        return mkErrResp("expected agent.session.started event mid-flight")
    }

    // wait for process to complete
    err = cmd.Wait()
    resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
    if err == nil {
        return resp, nil
    }
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        resp.ExitCode = exitErr.ExitCode()
        return resp, nil
    }
    if ctx.Err() != nil {
        return resp, ctx.Err()
    }
    return resp, err
}
```
