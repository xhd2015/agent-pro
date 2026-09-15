# Scenario

**Feature**: one `usage collect` run writes one snapshot per provider

```
agent-pro usage collect --grok-home H --codex-home H --commandcode-home H [--json]
  -> <home>/usages/<provider>/<today>/<HH-MM-SS>-snapshot.jsonl   x3
  -> exit 0
```

## Preconditions

- Inherits the built binary, the temp homes, the fixture usage and the fake
  Command Code API.
- The three `--*-home` flags select the homes, so the flag wiring itself is under
  test; `AGENT_PRO_HOME` comes from the environment.

## Steps

1. Set the command shape: `usage collect` with the three home flags.
2. Leaves add `--json`, drop a fixture, or drop credentials.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Command = []string{"usage", "collect"}
req.WithHomes = true
return nil
}
```
