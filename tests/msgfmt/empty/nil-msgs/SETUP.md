# Scenario

**Feature**: `nil` message slice formats to empty string

```
Format(nil, Options{}) -> ""
FormatDetailed(nil, Options{}) -> zero Result
```

## Preconditions

- `nil` and empty slice are equivalent for formatting.

## Steps

1. Set `req.Msgs = nil`.
2. Leave default zero `Options`.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = nil
	req.Opts = msgfmt.Options{}
	return nil
}
```
