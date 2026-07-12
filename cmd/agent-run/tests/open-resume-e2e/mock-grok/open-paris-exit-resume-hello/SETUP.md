# Scenario

**Feature**: full open→Paris→/exit→exited→resume --open hello with mock-grok

```
agent-run run --open --session-id test-open-resume-e2e \
  --agent-runner grok-tty \
  --agent-runner-binary llm-mock-run-grok \
  --agent-runner-config-home <grok-home> \
  --dir <workspace> \
  "one word of France capital"
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  + multi-turn LLM_MOCK_RUN_GROK_COMMAND
  -> wait snapshot/events contain "Paris"
  -> agent-run send test-open-resume-e2e /exit
  -> wait status runner.exited=true
  -> agent-run resume --open test-open-resume-e2e "hello"
  -> NOT already in use; snapshot/events show HELLO_RESUME_MARKER or hello UI
```

Mirrors `script/debug/open-resume-e2e` with mock instead of real grok.

## Steps

1. Pin session id, prompts, and scenario for the primary integration leaf.
2. Root/mock Setup already wired binary, homes, and mock hook.
3. Run orchestrates the multi-step flow; Assert checks all gates.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "open-paris-exit-resume-hello"
	req.SessionID = defaultSessionID
	req.GrokSessionUUID = defaultGrokSessionUUID
	req.OpenPrompt = defaultOpenPrompt
	req.WantParis = defaultWantParis
	req.FollowupPrompt = defaultFollowup
	req.HelloMarker = defaultHelloMarker
	// Re-apply mock env in case leaf constants differ from root defaults.
	configureMockGrokEnv(t, req)
	return nil
}
```
