## Expected

- `resp.Slug` is `codex_run_model_gpt-5.5_medium` (leading `--` stripped from flag tokens).
- `resp.ServeSubcommand` is `__serve_codex_run_model_gpt-5.5_medium__`.
- Slug contains only alnum, `-`, `.`, and `_` separators.

## Errors

- None from `Run`.

```go
import (
	"regexp"
	"testing"
)

var slugCharsetRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	wantSlug := "codex_run_model_gpt-5.5_medium"
	if resp.Slug != wantSlug {
		t.Fatalf("Slug = %q, want %q", resp.Slug, wantSlug)
	}
	if !slugCharsetRe.MatchString(resp.Slug) {
		t.Fatalf("slug contains disallowed characters: %q", resp.Slug)
	}
	assertServeWrap(t, resp.Slug, resp.ServeSubcommand)
	assertNoBareServe(t, resp.ServeSubcommand)
}
```