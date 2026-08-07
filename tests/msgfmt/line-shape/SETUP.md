# Scenario

**Feature**: single-message line shape and empty field omission

```
one Message + zero Options
  -> "Chat history (1 message):\n" + line + "\n"
```

## Preconditions

- Exactly one source message, no caps triggered.
- Header is the singular form `Chat history (1 message):`.
- Line omit rules from root DSN apply.

## Steps

1. Branch Setup clears selection/budget options (line-shape only).
2. Leaf builds one `Message` with the field combination under test.
3. Assert exact full `Text` (header + line + trailing newline).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// line-shape: single message, no MaxMessages / TotalBudget interaction.
	req.Opts = msgfmt.Options{}
	return nil
}
```
