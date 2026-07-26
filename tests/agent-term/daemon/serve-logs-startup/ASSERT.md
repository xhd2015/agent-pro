## Expected

- Daemon stderr contains the listen address within startup timeout.
- Message mentions `listening` and the bound host:port.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", resp.DaemonPort)
	assert.Output(t, resp.DaemonStderr, fmt.Sprintf(`
<contains>
listening
%s
</contains>`, addr))
}
```