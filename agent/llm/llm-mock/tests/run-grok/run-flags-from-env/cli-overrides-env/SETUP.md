# Scenario

**Feature**: explicit CLI `--log-events` on `llm-mock run` overrides env duplicate (last-wins)

```
LLM_MOCK_RUN_FLAGS=--log-events a.jsonl
llm-mock run --log-events b.jsonl grok -> only b.jsonl written
fake grok -> one curl -> one message AgentEvent in b.jsonl
```

## Steps

1. Config JSON with one exchange.
2. `RunFlagsEnv` points at `a.jsonl`; `LogEventsPath` is `b.jsonl` passed on CLI (`OmitCLIRunFlags` false).
3. Fake grok curls mock once.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	tmp := t.TempDir()
	req.ConfigEnv = "file"
	req.FakeGrokCmd = fakeGrokCurlOnce
	req.OmitCLIRunFlags = false
	req.ConfigJSON = minimalMockConfigJSON(8080, `[
    {
      "request": {"role": "user", "content": "config-only-prompt", "index": -1},
      "response": {"content": "from-config", "finish_reason": "stop"}
    }
  ]`)
	envPath := filepath.Join(tmp, "a.jsonl")
	cliPath := filepath.Join(tmp, "b.jsonl")
	req.RunFlagsEnv = "--log-events " + envPath
	req.LogEventsPath = cliPath
	req.EnvLogEventsOverridePath = envPath
	return nil
}
```