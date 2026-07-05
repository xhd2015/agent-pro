## Expected

- Exit code 0.
- Fake opencode output contains `OPENCODE_ARGV=run --model llm-mock/mock-model hi` (argv after `opencode` unchanged).
- `--log-events` path did not consume `run`, `--model`, or `hi`.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "OPENCODE_ARGV=run --model llm-mock/mock-model hi")
}
```