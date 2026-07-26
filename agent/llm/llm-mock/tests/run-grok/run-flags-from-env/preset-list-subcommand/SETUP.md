# Scenario

**Feature**: `LLM_MOCK_RUN_FLAGS=--mock-events-preset=list` lists presets on `llm-mock run grok` without starting grok

```
LLM_MOCK_RUN_FLAGS=--mock-events-preset=list
llm-mock run grok -> ParseRunFlags -> catalog stdout, exit 0
(no GROK_HOME=, grok not started)
```

## Steps

1. Set `RunFlagsEnv` to `--mock-events-preset=list`.
2. Omit CLI `--mock-events-preset`; pass `grok` subcommand on argv (`ListOnly` false).
3. Do not set fake grok hook — grok must not run.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.OmitCLIRunFlags = true
	req.RunFlagsEnv = "--mock-events-preset=list"
	req.ListOnly = false
	req.ExpectedExit = 0
	return nil
}
```