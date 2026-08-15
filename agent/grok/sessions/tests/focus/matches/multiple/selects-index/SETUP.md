# Scenario

**Feature**: requested candidate index is focused

```
focus --index 1 <id> -> candidates [0,1] -> focus candidate 1 only
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{req.SessionID, "--index", "1"}
	return nil
}
```
