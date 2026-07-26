# Scenario

**Feature**: markdown title meta labels and char annotations are stripped

```
# agent returns non-JSON: **Title (N chars):** `feat: ...`
fake-opencode -> sanitize strips meta + unwraps ticks -> clean title on stdout
```

## Preconditions
- Fixture `md_title_char_annotation`.

## Steps
1. Stage a change.
2. Mock agent text from fixture.
3. Run without `--commit`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	StageRepoWithChange(t, req)
	WriteMockAgentText(t, req, "sess_md_meta", ReadAntiPatternIn(t, "md_title_char_annotation"))
	req.Commit = false
	req.Operation = "md_title_char_annotation"
	return nil
}
```
