# Scenario

**Feature**: flag parsing stops at `grok`; preset flag does not consume grok argv

```
llm-mock run --mock-events-preset=simple grok -p hello
orchestrator -> grok CLI with argv "-p hello"
```

## Steps

1. Install fake `grok` on PATH (echoes `$*` as `GROK_ARGV=...`).
2. Pass `--mock-events-preset=simple` and grok args `-p hello`.
3. Do not use `LLM_MOCK_RUN_GROK_COMMAND` (hook ignores argv).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := installFakeGrokEchoArgv(t, req); err != nil {
		return err
	}
	req.MockEventsPreset = "simple"
	req.GrokArgs = []string{"-p", "hello"}
	return nil
}
```