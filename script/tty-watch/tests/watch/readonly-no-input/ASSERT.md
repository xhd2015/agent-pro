## Expected

- Watch output does not contain `SHOULD_NOT_ECHO` (stdin probe was not forwarded).
- Observer mode is readonly.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	combined := resp.Combined
	if combined == "" {
		combined = resp.Stdout
	}
	if strings.Contains(combined, "SHOULD_NOT_ECHO") {
		t.Fatalf("watch forwarded stdin to session, output %q", combined)
	}
}
```