# Scenario

**Feature**: status resolves sessions by ref form (compound runner/id or
`--grok-session-id`)

```
# compound (legacy leaf; bare id preferred)
seed meta
  -> agent-run status grok-tty/<id> -> multi-layer view for that session

# explicit Grok provider id (meta-only)
seed meta.runner in {grok, grok-tty} + meta.runner_session_id = UUID
  -> agent-run status --grok-session-id UUID
  -> resolve that SessionMeta; continue normal status probe
  # mutex with positional; 0/2+ matches fail; non-grok runners never match
```

## Steps

1. Leaf seeds unique meta and chooses ref form (`runner/session` or
   `--grok-session-id`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Grouping for session-ref form leaves (compound or --grok-session-id).
	if req.Runner == "" {
		req.Runner = defaultRunner
	}
	return nil
}
```
