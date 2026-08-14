# Scenario

**Feature**: Codex last glyph line holds user draft

```
last ›/» remainder non-empty AND line does not contain " medium · "
  -> DetectInputBox
  -> occupied
```

## Preconditions

- Footer may appear on a **following** line (live occupied / multiline).
- Sibling empty branch covers glue-on-same-line.

## Steps

1. Mark family `codex-occupied`.
2. Leaf injects live single-line, multiline, hint-without-glue, or `»` draft.

## Context

`sendable` stays idle on the live occupied shapes; that contract is asserted
under `sendable-independent/`, not here.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Family = "codex-occupied"
	return nil
}
```
