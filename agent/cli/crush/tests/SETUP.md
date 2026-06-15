## Preconditions
- The `crush` binary is available in PATH.
- Default mode (`Mode=""`) runs real queries via subprocess.
- `Mode="convert"` tests `UnwrapEvent` (no server needed).
- `Mode="server-client"` tests `CrushServerClient` operations (auto-detects `crush` in PATH; skips if not found).
- `Mode="server-ask"` tests `CrushAgent.Ask()` with server client.
- `Mode="convert-roundtrip"` tests `crush_types.ToCrush`/`crush_types.FromCrush` conversion round-trip (no server needed).

## Steps
1. Dispatch based on `Mode` field:
   - `""` → `runSubprocess`: look up crush binary, create CrushAgent, call Ask.
   - `"convert"` → `runConvert`: call `crush.UnwrapEvent`, return result as JSON.
   - `"server-client"` → `runServerClient`: run specified CrushServerClient operation.
   - `"server-ask"` → `runServerAsk`: create CrushAgent with server client, call Ask.
   - `"convert-roundtrip"` → `runConvertRoundtrip`: call `crush_types.ToCrush` then `crush_types.FromCrush`, return round-tripped events as JSON.

```go
import (
	"encoding/json"
	"fmt"
	osexec "os/exec"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/cli/crush"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/exec"
)

type Request struct {
	Prompt       string
	ResumePrompt string
	Model        string
	Env          []string

	Mode            string // "", "convert", "server-client", "server-ask", "convert-roundtrip"
	SSEInput        string // raw SSE data line for convert mode
	ServerOperation string // "health-check", "auto-start", "create-workspace", "send-and-receive", "server-lifecycle", "server-reuse"
	AgentEventsJSON string // JSON array of types.AgentEvent for convert-roundtrip mode
	SessionID       string // session ID for convert-roundtrip mode
}

type Response struct {
	Answer string // for Ask results
	Output string // for convert / server-client results (JSON)
}

func Setup(t *testing.T, req *Request) error {
	req.Model = ""
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Mode {
	case "convert":
		return runConvert(t, req)
	case "server-client":
		return runServerClient(t, req)
	case "server-ask":
		return runServerAsk(t, req)
	case "convert-roundtrip":
		return runConvertRoundtrip(t, req)
	default:
		return runSubprocess(t, req)
	}
}

func runSubprocess(t *testing.T, req *Request) (*Response, error) {
	paths := &exec.PathsConfig{
		RootDirName: ".crush-agent-test",
		DataDirName: "data",
		BinDirName:  "bin",
	}
	env := exec.NewEnv(paths, "CRUSH_AGENT_TEST_ROOT")

	crushPath, err := crush.FindAgentPath(env)
	if err != nil {
		t.Skip("crush not found in PATH; skip integration test")
		return &Response{}, nil
	}

	agent := &crush.CrushAgent{
		AgentPath:    crushPath,
		SettingsPath: "",
		Workspace:    t.TempDir(),
		Env:          env,
	}

	ctx := t.Context()
	opts := &registry.AskOptions{
		Model: req.Model,
	}
	answer, err := agent.Ask(ctx, req.Prompt, opts, nil)
	if err != nil {
		return &Response{Answer: answer}, err
	}

	if req.ResumePrompt != "" {
		sessionID := agent.LastSessionID
		if sessionID == "" {
			t.Fatalf("expected session ID to be captured from first query response:\n%s", answer)
		}
		opts.SessionID = sessionID
		answer2, err := agent.Ask(ctx, req.ResumePrompt, opts, nil)
		return &Response{Answer: answer2}, err
	}

	return &Response{Answer: answer}, nil
}

func runConvertRoundtrip(t *testing.T, req *Request) (*Response, error) {
	var events []types.AgentEvent
	if err := json.Unmarshal([]byte(req.AgentEventsJSON), &events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AgentEventsJSON: %w", err)
	}
	crushEvents := crush_types.ToCrush(events, req.SessionID)
	resultEvents := crush_types.FromCrush(crushEvents, req.SessionID)
	data, err := json.Marshal(resultEvents)
	if err != nil {
		return nil, err
	}
	return &Response{Output: string(data)}, nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func countCrushServerProcesses() int {
	cmd := osexec.Command("pgrep", "-x", "crush")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return 0
	}
	return strings.Count(trimmed, "\n") + 1
}

func runConvert(t *testing.T, req *Request) (*Response, error) {
	event, err := crush.UnwrapEvent([]byte(req.SSEInput))
	if err != nil {
		return nil, err
	}
	if event == nil {
		return &Response{Output: "null"}, nil
	}
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return &Response{Output: string(data)}, nil
}

func runServerClient(t *testing.T, req *Request) (*Response, error) {
	client, err := crush.NewCrushServerClient()
	if err != nil {
		return nil, err
	}
	ctx := t.Context()
	// server-lifecycle and server-reuse manage their own server lifecycle;
	// all other operations require an already-running server.
	if req.ServerOperation != "server-lifecycle" && req.ServerOperation != "server-reuse" {
		if err := client.EnsureServer(ctx); err != nil {
			return nil, err
		}
	}
	switch req.ServerOperation {
	case "health-check":
		status, err := client.HealthCheck(ctx)
		if err != nil {
			return nil, err
		}
		result, _ := json.Marshal(map[string]int{"status": status})
		return &Response{Output: string(result)}, nil
	case "auto-start":
		status, err := client.HealthCheck(ctx)
		if err != nil {
			return nil, err
		}
		started := client.ServerStarted()
		result, _ := json.Marshal(map[string]any{"status": status, "started": started})
		return &Response{Output: string(result)}, nil
	case "create-workspace":
		cwd := t.TempDir()
		id, err := client.CreateWorkspace(ctx, cwd)
		if err != nil {
			return nil, err
		}
		result, _ := json.Marshal(map[string]string{"id": id})
		return &Response{Output: string(result)}, nil
	case "send-and-receive":
		cwd := t.TempDir()
		workspaceID, err := client.CreateWorkspace(ctx, cwd)
		if err != nil {
			return nil, err
		}
		if err := client.InitAgent(ctx, workspaceID); err != nil {
			return nil, err
		}
		sessionID, err := client.CreateSession(ctx, workspaceID, "")
		if err != nil {
			return nil, err
		}
		eventCh, err := client.SubscribeEvents(ctx, workspaceID)
		if err != nil {
			return nil, err
		}
		if err := client.SendMessage(ctx, workspaceID, "hello", sessionID); err != nil {
			return nil, err
		}
		var events []crush_types.Event
		for dataLine := range eventCh {
			data := crush.ExtractSSEData(dataLine)
			if data == "" {
				continue
			}
			event, err := crush.UnwrapEvent([]byte(data))
			if err != nil {
				return nil, err
			}
			if event == nil {
				continue
			}
			events = append(events, *event)
			if event.Type == crush_types.EventRunComplete {
				break
			}
		}
		eventsJSON, _ := json.Marshal(events)
		return &Response{Output: string(eventsJSON)}, nil
	case "server-lifecycle":
		// Kill any existing server from previous tests so we start clean.
		osexec.Command("pkill", "-x", "crush").Run()
		time.Sleep(200 * time.Millisecond)
		beforeCount := countCrushServerProcesses()
		if err := client.EnsureServer(ctx); err != nil {
			return nil, err
		}
		afterEnsureCount := countCrushServerProcesses()
		healthStatus, healthErr := client.HealthCheck(ctx)
		osexec.Command("pkill", "-x", "crush").Run()
		time.Sleep(200 * time.Millisecond)
		afterKillCount := countCrushServerProcesses()
		_, healthAfterKillErr := client.HealthCheck(ctx)
		result, _ := json.Marshal(map[string]any{
			"before_count":         beforeCount,
			"after_ensure_count":   afterEnsureCount,
			"health_status":        healthStatus,
			"health_err":           errString(healthErr),
			"after_kill_count":     afterKillCount,
			"health_after_kill_err": errString(healthAfterKillErr),
		})
		return &Response{Output: string(result)}, nil
	case "server-reuse":
		clientA, err := crush.NewCrushServerClient()
		if err != nil {
			return nil, err
		}
		clientB, err := crush.NewCrushServerClient()
		if err != nil {
			return nil, err
		}
		beforeCount := countCrushServerProcesses()
		if err := clientA.EnsureServer(ctx); err != nil {
			return nil, err
		}
		afterACount := countCrushServerProcesses()
		if err := clientB.EnsureServer(ctx); err != nil {
			return nil, err
		}
		afterBCount := countCrushServerProcesses()
		healthA, healthAErr := clientA.HealthCheck(ctx)
		healthB, healthBErr := clientB.HealthCheck(ctx)
		result, _ := json.Marshal(map[string]any{
			"before_count":     beforeCount,
			"after_A_count":    afterACount,
			"after_B_count":    afterBCount,
			"health_A_status":  healthA,
			"health_A_err":     errString(healthAErr),
			"health_B_status":  healthB,
			"health_B_err":     errString(healthBErr),
		})
		return &Response{Output: string(result)}, nil
	}
	return nil, fmt.Errorf("unknown server operation: %s", req.ServerOperation)
}

func runServerAsk(t *testing.T, req *Request) (*Response, error) {
	client, err := crush.NewCrushServerClient()
	if err != nil {
		return nil, err
	}
	agent := &crush.CrushAgent{
		ServerClient: client,
		Workspace:    t.TempDir(),
	}
	ctx := t.Context()
	opts := &registry.AskOptions{
		Model: req.Model,
	}
	answer, err := agent.Ask(ctx, req.Prompt, opts, nil)
	if err != nil {
		return &Response{Answer: answer}, err
	}
	if req.ResumePrompt != "" {
		sessionID := agent.LastSessionID
		if sessionID == "" {
			t.Fatalf("expected session ID to be captured from first query response:\n%s", answer)
		}
		opts.SessionID = sessionID
		answer2, err := agent.Ask(ctx, req.ResumePrompt, opts, nil)
		return &Response{Answer: answer2}, err
	}
	return &Response{Answer: answer}, nil
}
```
