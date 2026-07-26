# Scenario

**Feature**: run events persist to `events.jsonl` under `AGENT_RUN_HOME`

```
agent-run run --json → stdout lines match sessions/<runner>/<id>/events.jsonl
```

## Preconditions

- `fake-codex` on PATH (inherited from root `SETUP.md`).
- Grouping `Setup` prefixes `req.Args` with `run --agent-runner fake-codex`.

## Steps

1. Run `agent-run run --json --agent-runner fake-codex "hi"`.
2. Locate `events.jsonl` under `AGENT_RUN_HOME/sessions/`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = append(req.Args, "--json", "hi")
	return nil
}
```