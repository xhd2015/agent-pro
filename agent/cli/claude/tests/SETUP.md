# Scenario

**Feature**: ClaudeAgent shells out to the `claude` binary headless and parses its stream-json output

```
# ClaudeAgent resolves claude binary, spawns it headless, streams NDJSON
Run -> ClaudeAgent.Ask -> claude -p <q> --output-format stream-json --verbose
# each native line is converted to canonical AgentEvent JSONL on RawLog
ClaudeAgent <- claude (system init, assistant blocks, user tool_result, result)
ClaudeAgent -> ClaudeEventWriter -> RawLog (AgentEvent JSONL)
# session id captured from any event carrying session_id
ClaudeAgent <- claude (session_id on system init + result)
```

## Preconditions
- The `claude` binary is available in PATH for the LLM-calling leaves.
- Tests run real queries against the claude CLI through ClaudeAgent.
- LLM-calling leaves are skipped (not failed) when `claude` is absent, or when
  `CLAUDE_SKIP_INTEGRATION=1`.
- `binary-not-found` never spawns claude (it points at a bogus path).

## Steps
1. Look up the `claude` binary from PATH using `claude.FindAgentPath`; skip if
   not installed (unless an explicit `AgentPath` is set on the request).
2. Initialize the ClaudeAgent with the resolved binary path.
3. Dispatch based on operation mode:
   - For `Ask()` leaves: execute the query via `agent.Ask()` and return the
     answer, capturing `agent.LastSessionID`.
   - For `ListModels()` leaf: call `agent.ListModels()` and return the model
     list as JSON.
4. For session-resume, capture `LastSessionID` from the first Ask and reuse it
   via `AskOptions.SessionID`.

## Context
- `claude` has no model-listing command; `ListModels` returns `nil, nil`.
- The `--verbose` flag is required for `--output-format stream-json`.

```go
import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/claude"
	"github.com/xhd2015/agent-pro/agent/cli/registry"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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

func runAsk(t *testing.T, agent *claude.ClaudeAgent, req *Request) (*Response, error) {
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

func runListModels(t *testing.T, agent *claude.ClaudeAgent) (*Response, error) {
	ctx := t.Context()
	models, err := agent.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	// Preserve nil-ness: claude's ListModels returns nil, nil (unsupported).
	// Appending to a nil slice keeps modelIDs nil when models is empty/nil,
	// so the list-models assertion (resp.Models == nil) holds.
	var modelIDs []string
	for _, m := range models {
		modelIDs = append(modelIDs, m.ID)
	}
	return &Response{Models: modelIDs}, nil
}
```
