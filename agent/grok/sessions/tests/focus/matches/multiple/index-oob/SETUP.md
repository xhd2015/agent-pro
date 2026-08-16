# Scenario

**Feature**: an out-of-range --index is fatal

```
focus --index 2 <id> -> two candidates -> error, no focus
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{req.SessionID, "--index", "2"}
	return nil
}
```
