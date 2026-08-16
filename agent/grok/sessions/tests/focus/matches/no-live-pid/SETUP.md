# Scenario

**Feature**: known session with no live grok process is not found

```
known session + no grok open-file hit -> not found
# iTerm is not listed
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeProjectFocusSession(t, req)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	req.ITerm = oneITermTab()
	return nil
}
```
