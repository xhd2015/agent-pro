# Scenario

**Feature**: `--auto-send-or-resume` honors `--detach` on run/resume; ignores on live send

```
MODE=run    + --detach -> create detached daemon
MODE=resume + --detach -> resume detach reopen
MODE=send   + --detach -> note: ignored; send proceeds
```

## Preconditions

- Auto flag long form: `--auto-send-or-resume`.
- Leaves seed missing / exited / live sessions accordingly.

## Steps

1. Grouping documents auto+detach class.
2. Leaves set MODE fixtures and `--detach`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	return nil
}
```
