# Scenario

**Feature**: empty input produces empty output

```
nil or len==0 msgs + any Options -> Format "" / zero Result
```

## Preconditions

- Empty means no messages to render: no header, no lines.
- Options are irrelevant when there is nothing to format (leaves may pass zero opts).

## Steps

1. Branch Setup clears Options to zero defaults.
2. Leaf sets `req.Msgs` to `nil` or empty non-nil slice.
3. Assert empty Text and zero Result fields.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// empty-input branch: zero options; leaves choose nil vs empty slice for Msgs.
	req.Opts = msgfmt.Options{}
	return nil
}
```
