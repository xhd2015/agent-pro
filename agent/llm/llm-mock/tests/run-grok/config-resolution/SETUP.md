# Scenario

**Factor**: config source — none (default empty), `LLM_MOCK_CONFIG_FILE`, or legacy `LLM_MOCK_CONFIG`

```
orchestrator -> mockconfig loader
  # file env wins over legacy; neither -> default {"exchanges": []}
mockconfig <- LLM_MOCK_CONFIG_FILE / LLM_MOCK_CONFIG / (none)
```

## Preconditions

- Config resolution is the highest-impact parameter for run-grok.
- Success leaves use fake grok (`LLM_MOCK_RUN_GROK_COMMAND`) and minimal mock config.
- `no-config-starts-grok` omits both config env vars and config file; expects exit 0 with fake grok.

## Steps

1. Grouping `Setup` sets default fake grok and minimal config JSON.
2. Leaf `Setup` chooses `ConfigEnv` mode (`file`, `legacy`, or neither).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FakeGrokCmd = fakeGrokPrintHome
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	return nil
}
```
