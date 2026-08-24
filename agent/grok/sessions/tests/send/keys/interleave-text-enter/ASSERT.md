## Expected

- Payload is CSI-Up + `pick` + `\n` + `tail`.
- NoSubmit true because `--enter` is in the sequence.

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
	want := "\x1b[Apick\ntail"
	if c.Text != want {
		t.Fatalf("Text = %q want %q", c.Text, want)
	}
	if !c.Opts.NoSubmit {
		t.Fatalf("NoSubmit want true, opts=%+v", c.Opts)
	}
	if c.Opts.NoCtrlU {
		t.Fatalf("NoCtrlU should stay false when text present, opts=%+v", c.Opts)
	}
}
```
