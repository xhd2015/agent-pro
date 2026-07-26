# Scenario

**Feature**: real grok headless receives mocked LLM response and records session events

```
orchestrator -> mock (capital of France -> Paris)
real grok -p "What is the capital of France?" -m mock-model -> Paris
integration <- events.jsonl turn_started model_id mock-model
```

## Steps

1. Mock config exchange: `"capital of France"` → `"Paris"`.
2. Run headless grok: `-p "What is the capital of France?" --always-approve -m mock-model`.

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
	req.GrokArgs = []string{
		"-p", "What is the capital of France?",
		"--always-approve",
		"-m", "mock-model",
	}
	return nil
}
```
