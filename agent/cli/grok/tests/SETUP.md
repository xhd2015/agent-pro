## Preconditions
- The `grok` binary is available in PATH (install via `curl -fsSL https://x.ai/cli/install.sh | bash`).
- Tests run real queries against the grok CLI through GrokAgent.
- Tests are skipped if grok not found in PATH.

## Steps
1. Look up the grok binary from PATH using `grok.FindAgentPath`; skip if not installed.
2. Initialize the GrokAgent with the resolved binary path.
3. Dispatch based on operation mode:
   - For `Ask()` leaves: execute query via `agent.Ask()` and return the answer, capturing `LastSessionID`.
   - For `ListModels()` leaf: call `agent.ListModels()` and return the model list as JSON.
4. For session-resume, capture `LastSessionID` from first Ask and reuse via `AskOptions.SessionID`.

```go
import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/grok"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

func Setup(t *testing.T, req *Request) error {
	// Validate that Operation is set; default to ask if unset
	if req.Operation == "" {
		req.Operation = OpAsk
	}
	return nil
}

func parseRawLog(buf bytes.Buffer) []json.RawMessage {
	var events []json.RawMessage
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		events = append(events, json.RawMessage(line))
	}
	return events
}

func runAsk(t *testing.T, agent *grok.GrokAgent, req *Request) (*Response, error) {
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
			t.Fatalf("expected session ID to be captured from first query")
		}
		rawLogBuf.Reset()
		opts.SessionID = sessionID
		answer2, err := agent.Ask(ctx, req.ResumePrompt, opts, nil)
		events2 := parseRawLog(rawLogBuf)
		return &Response{Answer: answer2, SessionID: agent.LastSessionID, Events: events2}, err
	}

	return &Response{Answer: answer, SessionID: agent.LastSessionID, Events: events}, nil
}

func runListModels(t *testing.T, agent *grok.GrokAgent) (*Response, error) {
	ctx := t.Context()
	models, err := agent.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	modelIDs := make([]string, len(models))
	for i, m := range models {
		modelIDs[i] = m.ID
	}
	return &Response{Models: modelIDs}, nil
}
```
