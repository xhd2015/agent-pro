# Scenario

**Feature**: live Codex 0.147 empty box is hint glued to the model footer

```
› Summarize recent commitsgpt-5.6-terra medium · /private/…
  -> DetectInputBox
  -> empty
```

## Preconditions

- Fixture `codex-0.147-empty-glued.txt` matches the locked experiment last line
  (no space before `gpt-`; ` medium · ` on the **same** line as `›`).

## Steps

1. Set `req.Fixture` to the empty-glued fixture.

## Context

Naive “non-empty text after › ⇒ occupied” would fail this leaf.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Fixture = fixtureCodexEmptyGlued
	return nil
}
```
