## Preconditions
- The `pi` binary is available in PATH.
- PI must have API keys configured in its own config (`~/.pi/agent/`).
- This runs real queries against the pi CLI through PiAgent.

## Steps
1. Look up the pi binary from PATH using `pi.FindAgentPath`; skip if not installed.
2. Initialize the PiAgent with the resolved binary path.
3. Execute the query via `agent.Ask()` and return the answer.

```go
import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/pi"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

type Request struct {
	Prompt       string
	ResumePrompt string
	Model        string
	Env          []string
}

type Response struct {
	Answer    string
	SessionID string
	Events    []json.RawMessage
}

func Setup(t *testing.T, req *Request) error {
	req.Model = os.Getenv("PI_MODEL")
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	paths := &exec.PathsConfig{
		RootDirName: ".pi-agent-test",
		DataDirName: "data",
		BinDirName:  "bin",
	}
	env := exec.NewEnv(paths, "PI_AGENT_TEST_ROOT")

	piPath, err := pi.FindAgentPath(env)
	if err != nil {
		t.Skip("pi not found in PATH; skip integration test")
		return &Response{}, nil
	}

	agent := &pi.PiAgent{
		AgentPath:    piPath,
		SettingsPath: "",
		Workspace:    t.TempDir(),
		Env:          env,
	}

	ctx := t.Context()
	var rawLogBuf bytes.Buffer
	opts := &registry.AskOptions{
		Model:  req.Model,
		RawLog: &rawLogBuf,
	}
	answer, err := agent.Ask(ctx, req.Prompt, opts, nil)
	events := parseRawLog(rawLogBuf)
	if err != nil {
		return &Response{Answer: answer, SessionID: agent.LastSessionID, Events: events}, err
	}

	if req.ResumePrompt != "" {
		sessionID := agent.LastSessionID
		if sessionID == "" {
			t.Fatalf("expected session ID to be captured from first query response:\n%s", answer)
		}
		opts.SessionID = sessionID
		answer2, err := agent.Ask(ctx, req.ResumePrompt, opts, nil)
		events2 := parseRawLog(rawLogBuf)
		return &Response{Answer: answer2, SessionID: agent.LastSessionID, Events: events2}, err
	}

	return &Response{Answer: answer, SessionID: agent.LastSessionID, Events: events}, nil
}

func parseRawLog(buf bytes.Buffer) []json.RawMessage {
	var events []json.RawMessage
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		line = line[:len(line)-1] // trim newline
		if line == "" {
			continue
		}
		events = append(events, json.RawMessage(line))
	}
	return events
}
```
