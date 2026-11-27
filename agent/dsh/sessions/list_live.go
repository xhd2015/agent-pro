package sessions

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

// LiveSession is one DSH session whose write lease is held by a live process.
// The lease is the open session.lock descriptor, which DSH keeps for the life
// of the write handle and which the kernel drops when that process exits.
type LiveSession struct {
	ID  string
	PID int
}

// ListLiveOpts drives ListLive. Nil hooks use the live process list and lsof.
// The probe never acquires the write lease: taking the flock would make this
// process the session owner.
type ListLiveOpts struct {
	ListProcs func() []procresolve.Proc
	Lsof      func(pid int) []string
	// Root keeps only writers whose lease lives under this home. Empty reports
	// every DSH writer on the machine, which is right for a process that owns
	// the default backend and wrong for one pinned to another home: the pinned
	// process would list sessions from a directory it cannot read. The filter is
	// applied to the lease paths, so a caller does not have to compare file
	// prefixes itself.
	Root string
}

// listLiveCacheTTL bounds how often the process-table scan runs. The live
// lease makes one process listing plus one lsof exec the whole cost of
// ListLive, and the orchestrator polls the listing that calls it every three
// seconds; a short cache drops those syscalls while keeping a finished session
// from lingering for more than one poll.
const listLiveCacheTTL = 4 * time.Second

var listLiveCache struct {
	sync.Mutex
	rows    []LiveSession
	root    string
	expires time.Time
}

func cachedListLive(now time.Time, root string) ([]LiveSession, bool) {
	listLiveCache.Lock()
	defer listLiveCache.Unlock()
	if listLiveCache.rows == nil || now.After(listLiveCache.expires) || listLiveCache.root != root {
		return nil, false
	}
	out := make([]LiveSession, len(listLiveCache.rows))
	copy(out, listLiveCache.rows)
	return out, true
}

func storeListLive(rows []LiveSession, root string, now time.Time) {
	kept := make([]LiveSession, len(rows))
	copy(kept, rows)
	listLiveCache.Lock()
	defer listLiveCache.Unlock()
	listLiveCache.rows = kept
	listLiveCache.root = root
	listLiveCache.expires = now.Add(listLiveCacheTTL)
}

func resetListLiveCache() {
	listLiveCache.Lock()
	defer listLiveCache.Unlock()
	listLiveCache.rows = nil
	listLiveCache.expires = time.Time{}
}

// ListLive returns DSH sessions that currently have session.lock open.
// One process may hold many leases. An iTerm tab is not required: dsh web
// keeps its leases without one.
//
// Unhooked calls reuse a result for listLiveCacheTTL. Callers that inject
// hooks always run the scan, so tests and custom probes stay exact.
func ListLive(opts *ListLiveOpts) ([]LiveSession, error) {
	if opts == nil {
		opts = &ListLiveOpts{}
	}
	root := strings.TrimSpace(opts.Root)
	cacheable := opts.ListProcs == nil && opts.Lsof == nil
	if cacheable {
		if rows, ok := cachedListLive(time.Now(), root); ok {
			return rows, nil
		}
	}
	rows, err := scanListLive(opts, root)
	if err != nil {
		return nil, err
	}
	if cacheable {
		storeListLive(rows, root, time.Now())
	}
	return rows, nil
}

func scanListLive(opts *ListLiveOpts, root string) ([]LiveSession, error) {
	listProcs := opts.ListProcs
	if listProcs == nil {
		listProcs = procresolve.ListLiveProcs
	}
	lsof := opts.Lsof
	if lsof == nil {
		lsof = procresolve.LiveLsof
	}

	var pids []int
	for _, p := range listProcs() {
		if !procresolve.IsDSHWriter(p.Cmd) {
			continue
		}
		if p.PID <= 0 {
			continue
		}
		pids = append(pids, p.PID)
	}

	filesByPID := map[int][]string{}
	if opts.Lsof != nil {
		for _, pid := range pids {
			filesByPID[pid] = lsof(pid)
		}
	} else if len(pids) > 0 {
		filesByPID = procresolve.LiveLsofMany(pids)
	}

	seen := map[string]int{}
	var ids []string
	for _, pid := range pids {
		for _, path := range filesByPID[pid] {
			// A writer outside the pinned home holds a lease for a session this
			// store cannot read; reporting it would make the board list a row
			// whose log is somewhere else.
			if root != "" && !pathWithinRoot(root, path) {
				continue
			}
			kind, sid, ok := procresolve.ParseSessionFromPath(path)
			if !ok || kind != "dsh" || sid == "" {
				continue
			}
			if _, exists := seen[sid]; exists {
				continue
			}
			seen[sid] = pid
			ids = append(ids, sid)
		}
	}
	sort.Strings(ids)
	out := make([]LiveSession, 0, len(ids))
	for _, id := range ids {
		out = append(out, LiveSession{ID: id, PID: seen[id]})
	}
	return out, nil
}

// pathWithinRoot reports whether a lease path belongs to one pinned home.
func pathWithinRoot(root, path string) bool {
	cleanRoot := strings.TrimRight(filepath.Clean(root), string(filepath.Separator))
	cleaned := filepath.Clean(path)
	return cleaned == cleanRoot || strings.HasPrefix(cleaned, cleanRoot+string(filepath.Separator))
}
