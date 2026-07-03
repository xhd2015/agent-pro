## Expected

- Exit code 0.
- Stdout contains pid (12345), port (127.0.0.1:12345), tty type (grok-tty), session id (session-1), start time, and tcp reachable status.

## Expected Output

```
<contains>
<start-with>
pid
</start-with>
<hint:pid>12345</hint:pid>
<start-with>
port
</start-with>
<hint:port>127.0.0.1:12345</hint:port>
<start-with>
tty type
</start-with>
<hint:ttytype>grok-tty</hint:ttytype>
<start-with>
session id
</start-with>
<hint:sid>session-1</hint:sid>
<start-with>
session file path
</start-with>
<start-with>
start time
</start-with>
<hint:time>2026-07-03</hint:time>
</contains>
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "pid", "12345", "port", "127.0.0.1:12345", "tty type", "grok-tty", "session id", "session-1", "session file path", "start time", "2026-07-03")
}
```
