## Expected

- Exit code 0.
- Fake grok output contains `GROK_ARGV=--always-approve`.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "GROK_ARGV=--always-approve")
}
```