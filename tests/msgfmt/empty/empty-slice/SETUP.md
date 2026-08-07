# Scenario

**Feature**: empty non-nil message slice formats to empty string

```
Format([]Message{}, Options{}) -> ""
```

## Preconditions

- Non-nil empty slice must not emit a header.

## Steps

1. Set `req.Msgs` to `[]msgfmt.Message{}`.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{}
	req.Opts = msgfmt.Options{}
	return nil
}
```
