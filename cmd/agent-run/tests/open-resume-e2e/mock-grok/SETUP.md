# Scenario

**Feature**: open/resume E2E uses llm-mock-run-grok (no real grok CLI)

```
--agent-runner grok-tty
  --agent-runner-binary <session>/llm-mock-run-grok
  --agent-runner-config-home <leaf>/.grok
  + LLM_MOCK_RUN_GROK_COMMAND multi-turn
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  -> deterministic Paris / hello; no network grok
```

## Preconditions

- Root Setup already built session binaries and allocated leaf homes.
- Grouping only documents the mock-grok backend choice (MECE sibling would be
  real-grok, intentionally omitted from this tree).

## Steps

1. Confirm mock binary path and config home are set on `req`.
2. Leaves specialize session id / scenario / asserts.

```go
import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if strings.TrimSpace(req.LLMMockRunGrok) == "" {
		return fmt.Errorf("LLMMockRunGrok empty; root Setup must build llm-mock-run-grok")
	}
	if _, err := os.Stat(req.LLMMockRunGrok); err != nil {
		return fmt.Errorf("llm-mock-run-grok missing: %w", err)
	}
	if strings.TrimSpace(req.GrokHome) == "" {
		return fmt.Errorf("GrokHome empty")
	}
	// Ensure still no real-grok hook override.
	stripEnvPrefix(req, envGrokTTYCommand+"=")
	req.Scenario = "mock-grok"
	return nil
}
```
