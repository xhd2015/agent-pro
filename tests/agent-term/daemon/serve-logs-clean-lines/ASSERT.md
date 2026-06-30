## Expected

- Startup log line contains `listening on <addr>` without duplicating addr on next line.
- Request log contains `POST` and `/api/terminal/sessions` without mashing into listen address.
- Stderr must not mash listen address into HTTP method (e.g. `7681GET`, `7681POST`).

```go
import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", resp.DaemonPort)
	log := resp.DaemonStderr
	for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
		if strings.Contains(log, addr+method) {
			t.Fatalf("daemon log mashed listen address into %s line: %q", method, log)
		}
	}
	assert.Output(t, log, fmt.Sprintf(`
<contains>
listening on %s
POST
/api/terminal/sessions
</contains>`, addr))
}
```