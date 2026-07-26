# Scenario

**Factor**: config source — none (default empty) vs `LLM_MOCK_CONFIG_FILE`

```
orchestrator -> mockconfig loader
  # file env -> config path; neither -> default {"exchanges": []}
mockconfig <- LLM_MOCK_CONFIG_FILE / (none)
```

## Preconditions

- Config resolution is a high-impact parameter for run-opencode.
- Success leaves use fake opencode (`LLM_MOCK_RUN_OPENCODE_COMMAND`) and minimal mock config when file env is set.
- `no-config-starts-opencode` omits config env and config file; expects exit 0 with fake opencode.

## Steps

1. Grouping `Setup` sets default fake opencode and minimal config JSON.
2. Leaf `Setup` chooses `ConfigEnv` mode (`file` or neither).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FakeOpencodeCmd = fakeOpencodePrintConfigDir
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	return nil
}
```