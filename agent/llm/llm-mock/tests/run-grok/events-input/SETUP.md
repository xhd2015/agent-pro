# Scenario

**Factor**: `LLM_MOCK_EVENTS_FILE` input — additional exchanges appended after config

```
mockconfig loader -> merge config exchanges + events JSONL
llm-mock HTTP server <- merged exchanges[]
```

## Preconditions

- `LLM_MOCK_EVENTS_FILE` is input JSONL (not output).
- Exchanges from events file append after config `exchanges[]`.
- Fake grok uses `GROK_MODELS_BASE_URL` to curl the mock for deterministic responses.

## Steps

1. Grouping `Setup` sets `ConfigEnv` to `file` and fake grok curl hooks.
2. Leaf `Setup` sets or omits `EventsJSONL`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	return nil
}
```
