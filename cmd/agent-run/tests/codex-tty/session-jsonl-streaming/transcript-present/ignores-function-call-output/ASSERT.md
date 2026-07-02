## Expected Output

```text
<contains>
JSONL_ASSISTANT_AFTER_FUNCTION_OUTPUT
</contains>
```

## Expected

- Exit code 0.
- Stdout contains the real assistant message.
- Stdout does not contain `JSONL_FUNCTION_OUTPUT_SHOULD_NOT_PRINT`.

## Side Effects

- Function-call output may be retained for diagnostics but is not an assistant message.

## Errors

- No error is expected.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `
<contains>
JSONL_ASSISTANT_AFTER_FUNCTION_OUTPUT
</contains>`)
	if strings.Contains(resp.Stdout, codexFunctionOutputText) {
		t.Fatalf("function_call_output should not be emitted as assistant stdout:\n%s", resp.Stdout)
	}
}
```
