## Preconditions
- The `crush` binary is available in PATH.
- This runs real queries against the crush CLI.

## Steps
1. Look up the crush binary from PATH; skip if not installed.
2. Initialize the CrushAgent with the resolved binary path.
3. Execute the query and return the answer.

```go
import (
	"time"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/crush"
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
}

func Setup(t *testing.T, req *Request) error {
	req.Model = ""
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
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
		return &Response{Answer: answer, SessionID: agent.LastSessionID}, err
	}

	if req.ResumePrompt != "" {
		sessionID := agent.LastSessionID
		if sessionID == "" {
			t.Fatalf("expected session ID to be captured from first query response:\n%s", answer)
		}
		opts.SessionID = sessionID
		answer2, err := agent.Ask(ctx, req.ResumePrompt, opts, nil)
		return &Response{Answer: answer2, SessionID: agent.LastSessionID}, err
	}

	return &Response{Answer: answer, SessionID: agent.LastSessionID}, nil
}
```
