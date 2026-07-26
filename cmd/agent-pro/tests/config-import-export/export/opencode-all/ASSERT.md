## Expected
- Zip file is created with all opencode entries under `opencode/` prefix.
- Each source file maps to its expected zip path.
- No errors returned.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertZipContains(t, req.ZipPath, []string{
		"opencode/auth.json",
		"opencode/settings.json",
		"opencode/opencode.jsonc",
		"opencode/plugins/my-plugin.ts",
		"opencode/skills/my-skill/SKILL.md",
	})
}
```
