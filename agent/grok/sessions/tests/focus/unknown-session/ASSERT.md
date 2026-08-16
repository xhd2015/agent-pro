## Expected

- Unknown session id returns `not found` and does not discover or focus.

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
