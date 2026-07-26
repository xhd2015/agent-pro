# Scenario

**Feature**: per-word grok thought streaming lines coalesce into think AgentEvents

## Preconditions
- Grok CLI emits per-word `thought` events during streaming.
- Each native line is converted to a separate AgentEvent in events.jsonl.

## Steps
1. Set `req.GrokLines` to six grok `thought` streaming lines with word deltas.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokLines = []string{
		`{"type":"thought","data":"The"}`,
		`{"type":"thought","data":" user"}`,
		`{"type":"thought","data":" wants"}`,
		`{"type":"thought","data":" me"}`,
		`{"type":"thought","data":" to"}`,
		`{"type":"thought","data":" act"}`,
	}
	return nil
}
```