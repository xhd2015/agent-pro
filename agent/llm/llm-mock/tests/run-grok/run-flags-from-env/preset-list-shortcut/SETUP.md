# Scenario

**Feature**: `LLM_MOCK_RUN_FLAGS=--mock-events-preset=list` lists presets on `llm-mock-run-grok` without starting grok

```
LLM_MOCK_RUN_FLAGS=--mock-events-preset=list
llm-mock-run-grok -> ParseRunFlags(env+argv) -> catalog stdout, exit 0
(no GROK_HOME=, grok not started)
```

## Steps

1. `UseShortcut` true; `RunFlagsEnv` supplies `--mock-events-preset=list`.
2. No grok argv; shortcut has no run-flag argv beyond env.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.UseShortcut = true
	req.OmitCLIRunFlags = true
	req.RunFlagsEnv = "--mock-events-preset=list"
	req.ExpectedExit = 0
	return nil
}
```