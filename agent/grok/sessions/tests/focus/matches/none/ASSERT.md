## Expected

- A live TTY with no iTerm match is not found and does not focus.

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
	if resp.ListITermCalls != 1 {
		t.Fatalf("ListITermCalls = %d, want 1", resp.ListITermCalls)
	}
	if len(resp.Focused) != 0 {
		t.Fatalf("Focused = %v, want none", resp.Focused)
	}
}
```
