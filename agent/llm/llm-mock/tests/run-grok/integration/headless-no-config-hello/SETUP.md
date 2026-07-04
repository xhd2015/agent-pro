# Scenario

**Feature**: `llm-mock run grok` with no config must answer a simple hello prompt promptly

```
no LLM_MOCK_CONFIG_FILE / LLM_MOCK_CONFIG -> default exchanges[]
real grok -p "hello" -m mock-model -> random fallback must return quickly
grok TUI/headless must show assistant text; events.jsonl must reach first_token
```

## Steps

1. Do not set mock config env vars (default empty prefix).
2. Run orchestrator with `WorkDir = RepoRoot` so the background mock inherits the repo cwd (reproduces user session).
3. Run headless grok: `-p "hello" --always-approve -m mock-model`.
4. Assert exit 0 within 30 seconds with visible assistant output.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.WorkDir = req.RepoRoot
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.ExpectParis = false
	req.ExpectMockModel = true
	req.ExpectAssistantReply = true
	req.ExecTimeout = 30 * time.Second
	req.GrokArgs = []string{
		"-p", "hello",
		"--always-approve",
		"-m", "mock-model",
	}
	return nil
}
```