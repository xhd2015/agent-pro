package procresolve

import "fmt"

// FindAncestorGrok walks startPID then each PPID via opts.ListProcs and
// returns the nearest grok session runner (IsGrokRunner). Missing start PID,
// a missing parent, or pid 0 yields ok=false and a zero Proc.
func FindAncestorGrok(startPID int, opts Options) (Proc, bool) {
	var procs []Proc
	if opts.ListProcs != nil {
		procs = opts.ListProcs()
	}
	byPID := make(map[int]Proc, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
	}

	seen := make(map[int]bool)
	for pid := startPID; pid != 0; {
		if seen[pid] {
			break
		}
		seen[pid] = true
		p, ok := byPID[pid]
		if !ok {
			return Proc{}, false
		}
		if IsGrokRunner(p.Cmd) {
			return p, true
		}
		pid = p.PPID
	}
	return Proc{}, false
}

// ResolveFromAncestors finds the nearest grok ancestor of startPID and
// delegates to ResolveFromPID so the session id comes from Lsof paths.
// A missing start PID is an error. No grok ancestor is a soft miss
// (Kind=none); descendant groks are not considered.
func ResolveFromAncestors(startPID int, opts Options) (*Result, error) {
	var procs []Proc
	if opts.ListProcs != nil {
		procs = opts.ListProcs()
	}
	found := false
	for _, p := range procs {
		if p.PID == startPID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("pid not found: %d", startPID)
	}

	anc, ok := FindAncestorGrok(startPID, opts)
	if !ok {
		return &Result{
			InputPID: startPID,
			Kind:     "none",
		}, nil
	}
	return ResolveFromPID(anc.PID, opts)
}
