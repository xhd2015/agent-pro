# Scenario

**Feature**: blocking Update available menu classification

```
menu fixture (1. Update now / 2. Skip / 3. Skip until…) -> IsBlockingUpdateMenu=true
```

## Preconditions

Fixtures `01` and `02` contain full menu options and `Press enter to continue`.

## Steps

1. Leaf picks default selection vs Skip-selected fixture.

## Context

Auto-Skip protocol only runs when blocking menu is detected.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	return nil
}
```
