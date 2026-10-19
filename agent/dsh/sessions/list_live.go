package sessions

import (
	"sort"

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
}

// ListLive returns DSH sessions that currently have session.lock open.
// One process may hold many leases. An iTerm tab is not required: dsh web
// keeps its leases without one.
func ListLive(opts *ListLiveOpts) ([]LiveSession, error) {
	if opts == nil {
		opts = &ListLiveOpts{}
	}
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
