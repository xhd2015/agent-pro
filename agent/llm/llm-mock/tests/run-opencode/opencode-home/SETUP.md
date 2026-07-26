# Scenario

**Factor**: isolated opencode env — default temp dirs vs explicit `LLM_MOCK_OPENCODE_HOME` / `LLM_MOCK_OPENCODE_CONFIG_DIR`

```
orchestrator -> set OPENCODE_CONFIG_DIR + OPENCODE_CONFIG_CONTENT (baseURL -> mock)
orchestrator -> set HOME + isolation env vars for opencode child
fake opencode <- OPENCODE_CONFIG_DIR path (printed or preset)
```

## Preconditions

- Orchestrator sets `OPENCODE_CONFIG_CONTENT` pointing LLM traffic at the mock server.
- Default paths are fresh temp dirs removed on exit.
- Explicit `LLM_MOCK_OPENCODE_HOME` / `LLM_MOCK_OPENCODE_CONFIG_DIR` use caller-provided directories.

## Steps

1. Grouping `Setup` enables config content post-check and fake opencode that prints `OPENCODE_CONFIG_DIR`.
2. Leaf `Setup` chooses default temp vs explicit home/config dir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeOpencodeCmd = fakeOpencodePrintConfigDir
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	req.ExpectConfig = true
	return nil
}
```