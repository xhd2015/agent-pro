## Expected

- Exit code 0.
- Fake codex output contains `CODEX_ARGV=exec -m mock-model hi` (argv after `codex` unchanged).
- `--log-http` path did not consume `exec`, `-m`, or `hi`.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "CODEX_ARGV=exec -m mock-model hi")
}
```