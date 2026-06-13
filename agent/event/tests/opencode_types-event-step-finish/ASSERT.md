## Expected
- JSON parses into Event correctly.
- All step_finish fields including nested tokens/cache are populated with correct values.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "type=step_finish")
	assertContains(t, resp.Stdout, "sessionID=sess_sf")
	assertContains(t, resp.Stdout, "timestamp=1718200000456")
	assertContains(t, resp.Stdout, "part.id=p2")
	assertContains(t, resp.Stdout, "part.type=step-finish")
	assertContains(t, resp.Stdout, "part.reason=stop")
	assertContains(t, resp.Stdout, "part.snapshot=snap_xyz")
	assertContains(t, resp.Stdout, "part.cost=0.015")
	assertContains(t, resp.Stdout, "tokens.input=120")
	assertContains(t, resp.Stdout, "tokens.output=80")
	assertContains(t, resp.Stdout, "tokens.reasoning=40")
	assertContains(t, resp.Stdout, "tokens.cache.read=10")
	assertContains(t, resp.Stdout, "tokens.cache.write=5")
}
```
