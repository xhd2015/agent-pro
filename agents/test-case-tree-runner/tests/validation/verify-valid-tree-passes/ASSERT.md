## Expected
- Compile succeeds (exit code 0)
- No error messages

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Passed {
		t.Fatalf("expected compile to pass, got output:\n%s", resp.Output)
	}
}
```
