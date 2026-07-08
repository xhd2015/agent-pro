## Expected

- First `TryAcquireSessionLock` returns `acquired=true`.
- Second `TryAcquireSessionLock` returns `acquired=false` (non-blocking skip).
- No error from either attempt.

## Errors

- Second acquire succeeding indicates missing `LOCK_NB` or lock not held.

## Exit Code

N/A (direct package call)

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.FirstLockAcquired {
		t.Fatal("first lock acquirer must succeed")
	}
	if resp.SecondLockAcquired {
		t.Fatal("second lock acquirer must fail non-blocking while holder active")
	}
}
```