# Scenario

**Feature**: real codex headless receives mocked LLM response via Responses API

```
orchestrator -> mock (capital of France -> Paris)
real codex exec --skip-git-repo-check -m mock-model "What is the capital of France?" -> Paris
integration <- stdout contains Paris
```

## Steps

1. Mock config exchange: `"capital of France"` → `"Paris"`.
2. Run headless codex: `exec --skip-git-repo-check -m mock-model "What is the capital of France?"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {
        "role": "user",
        "content": "capital of France",
        "index": -1
      },
      "response": {
        "content": "Paris",
        "finish_reason": "stop"
      }
    }
  ]
}`
	req.CodexArgs = []string{
		"exec",
		"--skip-git-repo-check",
		"-m", "mock-model",
		"What is the capital of France?",
	}
	return nil
}
```