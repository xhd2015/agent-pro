## Expected

- `resp.Slug` length is at most 128 characters (≈120 body + hash suffix).
- `resp.Slug` ends with an underscore + 8 lowercase hex hash (`_[0-9a-f]{8}`).
- `resp.Slug` starts with `codex_flag-00` prefix (truncation keeps head of joined argv).
- `resp.ServeSubcommand` matches `__serve_{slug}__` and is not the legacy `__serve__` token.
- Naive full join (without truncation rules) would exceed 120 characters — slug must be shorter.

## Errors

- None from `Run`.

```go
import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

var truncateHashSuffixRe = regexp.MustCompile(`_[0-9a-f]{8}$`)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Slug) > 128 {
		t.Fatalf("slug length %d exceeds 128-char cap, slug=%q", len(resp.Slug), resp.Slug)
	}
	if !truncateHashSuffixRe.MatchString(resp.Slug) {
		t.Fatalf("truncated slug must end with _<8-hex-hash>, got %q", resp.Slug)
	}
	if !strings.HasPrefix(resp.Slug, "codex_flag-00") {
		t.Fatalf("truncated slug should preserve argv head, got %q", resp.Slug)
	}
	naive := strings.Join(req.Argv, "_")
	for _, r := range []string{"/", "\\", ";", "|", "&"} {
		naive = strings.ReplaceAll(naive, r, "_")
	}
	if len(naive) <= 120 {
		t.Fatalf("fixture argv should exceed 120 chars naive join (got %d)", len(naive))
	}
	if len(resp.Slug) >= len(naive) {
		t.Fatalf("expected truncated slug shorter than naive join: slug=%d naive=%d", len(resp.Slug), len(naive))
	}
	assertServeWrap(t, resp.Slug, resp.ServeSubcommand)
	assertNoBareServe(t, resp.ServeSubcommand)
	if resp.ServeSubcommand == fmt.Sprintf("__serve_%s__", naive) {
		t.Fatal("ServeSubcommand must use truncated slug, not full naive join")
	}
}
```