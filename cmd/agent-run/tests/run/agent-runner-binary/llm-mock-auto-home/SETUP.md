# Scenario

**Feature**: `llm-mock-run-grok` auto-provisions shared grok home for discovery

```
agent-run run --agent-runner-binary=llm-mock-run-grok (no config home)
  -> auto GROK_HOME on child
  -> LLM_MOCK_RUN_GROK_COMMAND seeds updates.jsonl
  -> stderr grok session + streamed assistant (not scrollback GROK_HOME= pollution)
```

## Steps

1. Build `llm-mock-run-grok` into temp `bin/`.
2. Set `LLM_MOCK_RUN_GROK_COMMAND` hook that prints banner and seeds `updates.jsonl`.
3. Run without `--agent-runner-config-home`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const autoHomeUUID = "a1111111-1111-4111-8111-111111111111"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := buildLLMMockRunGrok(t, req); err != nil {
		return err
	}
	req.Prompt = "auto home probe"
	req.GrokSessionUUID = autoHomeUUID
	req.Env = append(req.Env, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+autoHomeUUID)
	// Build proves llm-mock-run-grok compiles; harness script (same basename) exercises
	// auto-provisioned GROK_HOME without starting the mock HTTP server.
	req.GrokHomeProbePath = filepath.Join(req.TempDir, "grok-home-probe.log")
	if err := writeLLMMockRunGrokHarness(t, req.LLMMockRunGrok, req.Prompt, autoHomeUUID, "AUTO_HOME_STREAM_MARKER", req.GrokHomeProbePath); err != nil {
		return err
	}
	req.AgentRunnerBinary = req.LLMMockRunGrok
	req.Args = append(req.Args, "--agent-runner-binary", req.AgentRunnerBinary, req.Prompt)
	return nil
}
```