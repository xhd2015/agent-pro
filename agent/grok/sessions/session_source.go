package sessions

import (
	"fmt"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// SessionSourceOpts drives ResolveSessionSource. Tab hooks mirror TabResolveOpts.
type SessionSourceOpts struct {
	ListProcs        func() []FocusProc
	Lsof             func(int) []string
	ListITerm        func() ([]iterm2.SessionRef, error)
	CurrentSessionID func() string
	ControllingTTY   func() string
	AncestorTTYs     func() []string
}

// ResolveSessionSource picks exactly one of: positional session id, --tab, or
// --tab-index. Shared by fork, open, and snapshot so tab→session always goes
// through ResolveFromTab.
func ResolveSessionSource(positional []string, tabFlag *string, tabIndexFlag *int, opts *SessionSourceOpts) (sessionID string, tabMeta *TabResolveResult, err error) {
	if opts == nil {
		opts = &SessionSourceOpts{}
	}
	tabSet := tabFlag != nil
	tabIndexSet := tabIndexFlag != nil
	if tabSet && tabIndexSet {
		return "", nil, fmt.Errorf("--tab and --tab-index cannot be specified together")
	}

	switch {
	case tabSet || tabIndexSet:
		if len(positional) > 0 {
			return "", nil, fmt.Errorf("session id cannot be combined with --tab/--tab-index")
		}
		var sel TabSelector
		if tabSet {
			sel, err = ParseTabFlag(*tabFlag)
		} else {
			sel, err = ParseTabIndexFlag(*tabIndexFlag)
		}
		if err != nil {
			return "", nil, err
		}
		tr, err := ResolveFromTab(sel, &TabResolveOpts{
			ListProcs:        opts.ListProcs,
			Lsof:             opts.Lsof,
			ListITerm:        opts.ListITerm,
			CurrentSessionID: opts.CurrentSessionID,
			ControllingTTY:   opts.ControllingTTY,
			AncestorTTYs:     opts.AncestorTTYs,
		})
		if err != nil {
			return "", nil, err
		}
		return tr.SessionID, tr, nil
	default:
		if len(positional) != 1 {
			return "", nil, fmt.Errorf("expected session id, or --tab / --tab-index")
		}
		sessionID = strings.TrimSpace(positional[0])
		if sessionID == "" {
			return "", nil, fmt.Errorf("session id is required")
		}
		return sessionID, nil, nil
	}
}
