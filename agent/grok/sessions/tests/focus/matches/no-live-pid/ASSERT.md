## Expected

- No live grok PID returns `not found` without listing iTerm.

## Expected Output

```text
not found
```

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNotFound(t, resp)
	assertNoITerm(t, resp)
}
```
