# Scenario

**Feature**: large TotalBudgetRunes does not drop messages

```
3 short msgs, TotalBudgetRunes=10_000
  -> showing all 3 of 3; all bodies present
```

## Preconditions

- Budget is only a ceiling; large values behave like no budget.

## Steps

1. Three text-only messages; budget 10000.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{
		msg("", "", "alpha"),
		msg("", "", "beta"),
		msg("", "", "gamma"),
	}
	req.Opts = msgfmt.Options{TotalBudgetRunes: 10000}
	return nil
}
```
