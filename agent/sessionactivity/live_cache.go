package sessionactivity

import (
	"sync"
	"time"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

const dshLiveTTL = time.Second

var dshLive = struct {
	mu    sync.Mutex
	at    time.Time
	ids   map[string]bool
	known bool
}{}

func dshLiveIDs() (map[string]bool, bool) {
	dshLive.mu.Lock()
	defer dshLive.mu.Unlock()
	if dshLive.ids != nil && time.Since(dshLive.at) < dshLiveTTL {
		return dshLive.ids, dshLive.known
	}
	rows, err := dshsessions.ListLive(nil)
	dshLive.at = time.Now()
	if err != nil {
		dshLive.ids = map[string]bool{}
		dshLive.known = false
		return dshLive.ids, false
	}
	ids := make(map[string]bool, len(rows))
	for _, row := range rows {
		ids[row.ID] = true
	}
	dshLive.ids = ids
	dshLive.known = true
	return ids, true
}
