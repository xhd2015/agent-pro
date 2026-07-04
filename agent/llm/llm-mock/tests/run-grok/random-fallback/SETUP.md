# Scenario

**Profile**: fake grok curls mock HTTP — multi-turn no-config random fallback

```
no config -> default exchanges[]
fake grok -> curl mock 3x (turn 1 think+message, turn 2 multi-message)
orchestrator must keep mock alive; second user turn must not 400 no_match
```

## Preconditions

- Uses `LLM_MOCK_RUN_GROK_COMMAND` fake grok (not real grok).
- No `LLM_MOCK_CONFIG_FILE` / `LLM_MOCK_CONFIG` (default empty prefix).

## Steps

1. Grouping `Setup` ensures fake-grok profile (no real grok required).
2. Leaf `Setup` sets fake curl script and assertions on combined output.
3. `Run` starts orchestrator; fake grok hits mock like grok multi-turn session.
4. `Assert` checks all curls succeed without `no_match`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	return nil
}
```