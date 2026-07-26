## Expected

- After five create-on-connect + close-1000 cycles, `RunningProcessCount == 0`.
- No orphan interactive shells may remain holding PTYs.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.RunningProcessCount != 0 {
		t.Fatalf("create-on-connect churn left %d orphan shell process(es) (want 0); session_count=%d last_session=%s",
			resp.RunningProcessCount, resp.SessionCount, resp.SessionID)
	}
}
```
