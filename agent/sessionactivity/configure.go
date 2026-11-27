package sessionactivity

import (
	"strings"
	"sync"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

// Options pin the DSH home the activity folds read. The zero value keeps the
// default resolution ($DSH_HOME, then ~/.dsh), which is what a process that owns
// the default backend wants.
type Options struct {
	// Home is the DSH data directory. Empty means the default home.
	Home string
}

var (
	optionsMu sync.RWMutex
	options   Options
)

// Configure changes the home the activity folds read.
//
// A process that serves one pinned backend calls this before it starts handlers:
// the package builds its own store, so without it Read/Follow/WaitAfter would
// read the default home while the rest of the process reads the pinned one. The
// alternative, setting DSH_HOME in the environment, works only because every
// reader resolves the same variable; it also hides the choice from anyone reading
// the call.
//
// The per-session tail cache describes the previous home's logs and is dropped,
// so a changed home cannot answer from a cursor built against another file. This
// is a boot-time setting, not a hot swap: a change while reads are in flight is
// not coordinated with them.
func Configure(next Options) {
	next.Home = strings.TrimSpace(next.Home)
	optionsMu.Lock()
	changed := options.Home != next.Home
	options = next
	optionsMu.Unlock()
	if !changed {
		return
	}
	dshSessions.Range(func(key, _ any) bool {
		dshSessions.Delete(key)
		return true
	})
}

// storeOptions is the DSH home the folds read, as one value.
func storeOptions() dshsessions.Options {
	optionsMu.RLock()
	defer optionsMu.RUnlock()
	return dshsessions.Options{Home: options.Home}
}
