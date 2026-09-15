# Scenario

**Feature**: `--cron` repeats the collect cycle and `--max-runs` bounds it

```
usage collect --cron '@every 100ms' --max-runs 3
cycle now, then two more per schedule -> 3 records per provider
stderr: "usage collect: next run at <UTC>"   exit 0
```

## Preconditions

- `@every <duration>` is the interval form of the spec, so the leaf finishes in
  a few hundred milliseconds.
- Every provider answers, so any record count difference comes from the loop, not
  from failures.

## Steps

1. Run three cycles, 100ms apart, with JSON output so the cycles are countable.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Args = []string{"--cron", "@every 100ms", "--max-runs", "3", "--json"}
return nil
}
```
