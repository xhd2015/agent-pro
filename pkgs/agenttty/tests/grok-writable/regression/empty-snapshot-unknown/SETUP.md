# Scenario

**Feature**: empty boot snapshot yields unknown writable state

```
zero-length snapshot text
  -> CheckWritable
  -> ready=false, state=unknown, reason=no terminal output
```

## Preconditions

- Fixture `grok-boot-unknown-empty-e3b0c442.txt` is zero bytes (probe boot phase).

## Steps

1. Set `req.FixtureFile` to empty boot fixture.

## Context

- F4 boot guard; distinct from idle-with-prompt outcomes.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureBootEmpty
	return nil
}
```