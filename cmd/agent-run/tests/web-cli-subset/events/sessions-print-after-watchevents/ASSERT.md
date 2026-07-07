## Expected

- CLI stdout includes appended assistant line from tail.
- Exit code 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("sessions --print exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "WatchEvents appended line") {
		t.Fatalf("sessions --print did not tail appended event; stdout=%s", resp.Stdout)
	}
}
```
