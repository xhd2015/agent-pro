# Scenario

**Feature**: live Codex 0.147 occupied box is draft on the › line

```
› EXP-DRAFT-NOTE-42
  gpt-5.6-terra medium · /private/…
  -> DetectInputBox
  -> occupied
```

## Preconditions

- Fixture `codex-0.147-occupied-single.txt` matches the locked experiment:
  draft alone on the `›` line; footer on the next line.

## Steps

1. Set `req.Fixture` to the occupied-single fixture.

## Context

Footer still present in the snapshot — it must not count because it is not on
the last glyph line.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Fixture = fixtureCodexOccupiedSingle
	return nil
}
```
