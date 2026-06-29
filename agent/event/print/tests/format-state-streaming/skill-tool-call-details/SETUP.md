# Scenario

**Feature**: skill `tool_call` events show name and arguments below the header

## Preconditions
- **Reproduces**: maintain-topic `--print` follows `events.jsonl` through
  `FormatState.FormatLine`, which currently prints bare `SKILL` headers for
  canonical `tool_call` skill events.
- Session `20260625-114542-...-credit.pricing.center` recorded two consecutive
  skill invocations with only `tool_input.name`.

## Steps
1. Set `req.Lines` to the two skill `tool_call` events from that session.
2. Feed each line through `FormatState.FormatLine` like `FollowEventLog` does.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Lines = []string{
		`{"id":"prt_efce251440017YEGmef4U2QCm3","type":"tool_call","timestamp":1782359150956,"tool":"skill","tool_input":{"name":"confluence-fetch"},"mock":{"exit_code":0}}`,
		`{"id":"prt_efce25189001OoJCkuXbWK8Fty","type":"tool_call","timestamp":1782359151084,"tool":"skill","tool_input":{"name":"git-fetch"},"mock":{"exit_code":0}}`,
	}
	return nil
}
```