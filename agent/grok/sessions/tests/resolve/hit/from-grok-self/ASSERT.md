## Expected

- Session id from self; start_pid and ancestor_pid are both the grok pid.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	assertJSONDetails(t, resp.Stdout, ResolveDetailsExpect{
		SessionID:   fixtureSessionID,
		StartPID:    pidGrok,
		AncestorPID: pidGrok,
		RunnerPID:   pidGrok,
	})
}
```
