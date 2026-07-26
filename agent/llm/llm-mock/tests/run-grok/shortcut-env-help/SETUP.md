# Scenario

**Feature**: `LLM_MOCK_RUN_FLAGS=--help` prints run help on shortcut without grok/mock

```
LLM_MOCK_RUN_FLAGS=--help
llm-mock-run-grok -> ParseRunFlagsFromEnv -> help stdout, exit 0
(no GROK_HOME= stderr from orchestrator)
```

## Steps

1. `UseShortcut` true; `RunFlagsEnv` supplies `--help`.
2. No grok argv; shortcut has no run-flag argv beyond env.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.UseShortcut = true
	req.OmitCLIRunFlags = true
	req.RunFlagsEnv = "--help"
	req.ExpectedExit = 0
	return nil
}
```