## Expected

- JSON includes session_id and the same detail fields as `-v`.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	assertJSONDetails(t, resp.Stdout, ResolveDetailsExpect{
		SessionID:   fixtureSessionID,
		StartPID:    pidStart,
		AncestorPID: pidGrok,
		RunnerPID:   pidGrok,
	})
	if resp.Stderr != "" {
		t.Fatalf("stderr want empty, got %q", resp.Stderr)
	}
}
```
