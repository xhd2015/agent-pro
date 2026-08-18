# Scenario

**Feature**: parse-only idle flags (no TTY)

```
ParseRunIdle(exitOnIdle, timeoutRaw) -> enabled / duration / error
CLI presentation -> Error: … on stderr, exit 1 when invalid
```

## Preconditions

- `agentruncli.ParseRunIdle` is the L2 parse seam. Do **not** call `Handle`
  (would start a keep-alive TTY if the duration parsed as a raw string).

## Steps

1. Grouping sets `Op=parse`.
2. Invalid vs valid branches set the expected outcome family.
3. Leaves set `ExitOnIdle` and `IdleTimeoutRaw`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opParse
	return nil
}
```
