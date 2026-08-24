## Expected

- SendText once with `"\x03"`; NoCtrlU and NoSubmit true; Focus false.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if len(resp.SendCalls) != 1 {
		t.Fatalf("SendCalls = %#v", resp.SendCalls)
	}
	c := resp.SendCalls[0]
	if c.Text != "\x03" {
		t.Fatalf("Text = %q", c.Text)
	}
	if !c.Opts.NoCtrlU || !c.Opts.NoSubmit || c.Opts.Focus {
		t.Fatalf("opts = %+v", c.Opts)
	}
}
```
