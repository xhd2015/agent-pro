# Scenario

**Feature**: real opencode headless receives mocked LLM response via Chat Completions API

```
orchestrator -> mock (capital of France -> Paris)
real opencode run "What is the capital of France?" --model llm-mock/mock-model --dir <workdir> -> Paris
integration <- combined output contains Paris; HTTP log proves mock model request
```

## Steps

1. Mock config exchange: `"capital of France"` → `"The capital of France is Paris."`.
2. Run headless opencode with `--log-http` for request recording proof.
3. Use temp workdir via `--dir`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "content": "capital of France",
        "index": -1
      },
      "response": {
        "content": "The capital of France is Paris.",
        "finish_reason": "stop"
      }
    }
  ]
}`
	req.LogHTTPPath = filepath.Join(t.TempDir(), "http.jsonl")
	req.OpencodeArgs = []string{
		"run",
		"What is the capital of France?",
		"--model", "llm-mock/mock-model",
		"--dir", req.WorkDir,
	}
	return nil
}
```