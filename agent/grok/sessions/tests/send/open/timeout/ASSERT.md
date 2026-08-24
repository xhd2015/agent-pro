## Expected

- Wait expires → timeout error; opener called; no successful SendText.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "timed out waiting for hosting tab")
	if len(resp.Opened) != 1 {
		t.Fatalf("Opened = %v, want 1 (resume attempted)", resp.Opened)
	}
	assertNoSend(t, resp)
	if resp.SleepCalls < 1 {
		t.Fatalf("SleepCalls = %d, want >= 1", resp.SleepCalls)
	}
}
```
