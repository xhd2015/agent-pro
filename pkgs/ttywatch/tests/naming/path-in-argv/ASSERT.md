## Expected

- `resp.Slug` is `usr_local_bin_codex_C_Program_Files_Grok_grok.exe`.
- Slug contains no `/` or `\` characters.
- No doubled underscores (`__`) in slug.
- `resp.ServeSubcommand` wraps slug as `__serve_{slug}__`.

## Errors

- None from `Run`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	wantSlug := "usr_local_bin_codex_C_Program_Files_Grok_grok.exe"
	if resp.Slug != wantSlug {
		t.Fatalf("Slug = %q, want %q", resp.Slug, wantSlug)
	}
	if strings.ContainsAny(resp.Slug, `/\`) {
		t.Fatalf("slug must not contain path separators: %q", resp.Slug)
	}
	if strings.Contains(resp.Slug, "__") {
		t.Fatalf("slug must collapse repeated underscores: %q", resp.Slug)
	}
	assertServeWrap(t, resp.Slug, resp.ServeSubcommand)
	assertNoBareServe(t, resp.ServeSubcommand)
}
```