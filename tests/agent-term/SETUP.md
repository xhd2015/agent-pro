# Scenario

**Feature**: agent-term CLI drives ptywrap TCP daemon

```
# CLI flow
agent-term serve -> daemon TCP -> list/run/attach/web subcommands
```

## Preconditions

- `cmd/agent-term` exists (implementer adds during feature work).
- Implementer provides `AGENT_TERM_SERVER` env or `--server` flag for client subcommands.
- `probeAttachByName` helper completes WS handshake without full TTY for attach-by-name leaf.

## Steps

1. Build `agent-term` binary to temp path.
2. Set `req.AgentTermBin`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AgentTermBin = buildAgentTerm(t)
	return nil
}
```