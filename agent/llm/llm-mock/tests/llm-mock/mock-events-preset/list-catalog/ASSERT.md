## Expected

- Exit code 0.
- Stdout contains every MVP preset name: `simple`, `think-message`, `multi-think`, `tool-bash`, `tool-read`, `think-tool-message`.
- No listening port announced (catalog-only; server did not start).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	combined := resp.Stdout + resp.Stderr
	for _, name := range []string{
		"simple",
		"think-message",
		"multi-think",
		"tool-bash",
		"tool-read",
		"think-tool-message",
	} {
		assertContains(t, combined, name)
	}

	if resp.Port != 0 {
		t.Fatalf("catalog list must not start server; got port %d", resp.Port)
	}
}
```