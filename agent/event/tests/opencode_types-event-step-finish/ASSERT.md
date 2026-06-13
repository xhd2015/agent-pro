## Expected
- JSON parses into Event correctly.
- All step_finish fields including nested tokens/cache are populated with correct values.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "type=step_finish")
	assertContains(t, resp.Output, "sessionID=sess_sf")
	assertContains(t, resp.Output, "timestamp=1718200000456")
	assertContains(t, resp.Output, "part.id=p2")
	assertContains(t, resp.Output, "part.type=step-finish")
	assertContains(t, resp.Output, "part.reason=stop")
	assertContains(t, resp.Output, "part.snapshot=snap_xyz")
	assertContains(t, resp.Output, "part.cost=0.015")
	assertContains(t, resp.Output, "tokens.input=120")
	assertContains(t, resp.Output, "tokens.output=80")
	assertContains(t, resp.Output, "tokens.reasoning=40")
	assertContains(t, resp.Output, "tokens.cache.read=10")
	assertContains(t, resp.Output, "tokens.cache.write=5")
}
```
