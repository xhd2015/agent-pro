## Expected

- Focus / NoSubmit / NoCtrlU plumbed into SendText.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.SendCalls) != 1 {
		t.Fatalf("SendCalls = %v, want 1", resp.SendCalls)
	}
	o := resp.SendCalls[0].Opts
	if !o.Focus || !o.NoSubmit || !o.NoCtrlU {
		t.Fatalf("opts = %+v, want Focus+NoSubmit+NoCtrlU", o)
	}
}
```
