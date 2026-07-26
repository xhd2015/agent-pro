# Scenario

**Feature**: legacy default layout creates standard nested session artifacts

```
# questions/, progress/, messages.jsonl under date/sess_* dir
subagent.Run(zero layout) -> nested sess dir with all default artifacts
```

## Preconditions

- Inherited legacy zero-value `SessionLayout`.

## Steps

1. Inherit `configureLegacyBase` from `legacy/` grouping.

## Context

- Descendant leaf asserts nested `sess_*` path and default artifact set.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.SessionID == "" {
		configureLegacyBase(t, req)
	}
	return nil
}```
