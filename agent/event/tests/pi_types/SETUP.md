# Scenario

**Feature**: The `pi_types` package must exist under `agent/event/pi_types` with `types.go` and `convert.go`

## Preconditions

- The `pi_types` package must exist under `agent/event/pi_types` with `types.go` and `convert.go`.
- `types.AgentEvent` has a `Phase EventPhase` field.
- `ToPi` converts `[]types.AgentEvent` to `[]pi_types.Event`.
- `FromPi` converts `[]pi_types.Event` to `[]types.AgentEvent`.

## Steps

1. For `Target="wire"`: unmarshal `JSONInput` into a `pi_types.Event`, then marshal back to JSON.
2. For `Target="to_pi"`: call `pi_types.ToPi(Events)` and marshal the result.
3. For `Target="from_pi"`: call `pi_types.FromPi(PiEvents)` and marshal the result.
4. Default: pass through `Output`.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = assertContains
	_ = assertNotContains
	return nil
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got string, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}
```
