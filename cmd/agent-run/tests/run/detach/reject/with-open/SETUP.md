# Scenario

**Feature**: `--detach` and `--open` are mutually exclusive

```
run|resume --detach --open … -> exit ≠ 0
```

## Steps

1. Grouping marks open conflict class.
2. Leaves use run vs resume surfaces.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Open conflict class: leaves finalize --detach + --open on run or resume.
	if len(req.Args) == 0 {
		req.Args = []string{"run"}
	}
	return nil
}
```
