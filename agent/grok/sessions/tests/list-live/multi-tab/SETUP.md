# Scenario

```go
import "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	addLiveGrokHost(req, 5001, "ttys148", fixtureListLiveSID, "3", 1)
	// Second tab same TTY → DiscoverFocusHosting returns both refs.
	req.ITerm = append(req.ITerm, iterm2.SessionRef{
		WindowID:   "3",
		WindowName: "work",
		TabIndex:   2,
		SessionID:  "iterm-ttys148-b",
		TTY:        "/dev/ttys148",
	})
	req.Args = []string{}
	return nil
}
```
