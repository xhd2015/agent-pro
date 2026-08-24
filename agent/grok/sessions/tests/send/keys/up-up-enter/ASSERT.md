## Expected

- Payload CSI-Up×2 + `\n`; key-only forces NoCtrlU+NoSubmit.

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
	want := "\x1b[A\x1b[A\n"
	if c.Text != want {
		t.Fatalf("Text = %q want %q", c.Text, want)
	}
	if !c.Opts.NoCtrlU || !c.Opts.NoSubmit {
		t.Fatalf("opts = %+v", c.Opts)
	}
}
```
