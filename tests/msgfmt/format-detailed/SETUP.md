# Scenario

**Feature**: `FormatDetailed` Result metadata and parity with `Format`

```
FormatDetailed(msgs, opts) -> Result
Format(msgs, opts) == Result.Text
```

## Preconditions

- `SourceCount` is always `len(input msgs)` (before selection).
- `OldestDropped == SourceCount - Shown`.
- `LastMessageID` is the newest **input** message's ID (not affected by drops of older ids).
- `BodiesTruncated` counts messages whose bodies were shortened by the cap.

## Steps

1. Branch Setup starts from zero Options; leaves stress specific Result fields.
2. Leaf configures msgs/opts for one metadata scenario.
3. Assert metadata (and usually exact Text).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// format-detailed branch: leaves own full Options for metadata scenarios.
	req.Opts = msgfmt.Options{}
	return nil
}
```
