# Scenario

**Feature**: real busy detection when agent is thinking in prompt region

```
prompt box visible with ❯
prompt tail contains "thinking"
  -> CheckWritable returns ready=false, state=busy
```

## Preconditions

- Fixture from `writable_test.go` `TestCheckGrokWritable_busyWhenThinking`.

## Steps

1. Set `req.FixtureFile` to busy-thinking regression fixture.

## Context

- Guards against over-scoping busy fix (F3); must remain busy after hardening.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureBusyThinking
	return nil
}
```