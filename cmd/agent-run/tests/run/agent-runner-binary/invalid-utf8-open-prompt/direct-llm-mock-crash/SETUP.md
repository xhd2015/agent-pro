# Scenario

**Leaf A**: direct `llm-mock-run-grok` (no agent-run) with invalid UTF-8 argv

```
llm-mock-run-grok <invalid-utf8 PROMPT>
  -> starts real grok with that argv
  -> real grok panics at std::env::args (env.rs) FAST
  -> must exit with panic within 3s
  -> still running after 3s => FAIL
```

## Steps

1. Override exec binary to built `llm-mock-run-grok` (not agent-run).
2. Args = positional PROMPT only (same incident body).
3. Timeout 3s.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Parent already built mock + set invalid Prompt + 3s budget.
	// Run the mock binary directly — no agent-run wrap.
	req.AgentRun = req.LLMMockRunGrok
	req.Args = []string{req.Prompt}
	req.ExecTimeout = invalidUTF8Budget
	return nil
}
```
