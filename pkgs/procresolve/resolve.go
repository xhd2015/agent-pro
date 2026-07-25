package procresolve

import (
	"fmt"
	"sort"
)

// ResolveFromPID builds the descendant tree of pid, classifies nodes, and
// resolves a hard session id from open files on grok/codex runner candidates.
// Session ids come from open-file paths only (not cmdline flags).
func ResolveFromPID(pid int, opts Options) (*Result, error) {
	var procs []Proc
	if opts.ListProcs != nil {
		procs = opts.ListProcs()
	}

	byPID := make(map[int]Proc, len(procs))
	children := make(map[int][]int, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
		children[p.PPID] = append(children[p.PPID], p.PID)
	}

	input, ok := byPID[pid]
	if !ok {
		return nil, fmt.Errorf("pid not found: %d", pid)
	}

	maxDepth := opts.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 16
	}

	// BFS descendants from input, capped by MaxDepth (depth of input = 0).
	type item struct {
		pid   int
		depth int
	}
	tree := make([]ProcNode, 0)
	depthOf := map[int]int{pid: 0}
	queue := []item{{pid: pid, depth: 0}}
	seen := map[int]bool{pid: true}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		p := byPID[cur.pid]

		role := classifyCmd(p.Cmd)
		if cur.pid == pid {
			// Requested root is always reported as "input".
			role = "input"
		}
		tree = append(tree, ProcNode{
			PID:  p.PID,
			PPID: p.PPID,
			Role: role,
			Cmd:  p.Cmd,
		})

		if cur.depth >= maxDepth {
			continue
		}
		// Stable child order by PID for deterministic trees.
		kids := append([]int(nil), children[cur.pid]...)
		sort.Ints(kids)
		for _, cpid := range kids {
			if seen[cpid] {
				continue
			}
			if _, exists := byPID[cpid]; !exists {
				continue
			}
			seen[cpid] = true
			d := cur.depth + 1
			depthOf[cpid] = d
			queue = append(queue, item{pid: cpid, depth: d})
		}
	}

	// Collect runner candidates (grok|codex by cmd), prefer deeper leaves.
	type cand struct {
		pid   int
		cmd   string
		kind  string
		depth int
		leaf  bool
	}
	var cands []cand
	for _, n := range tree {
		kind := runnerKind(n.Cmd)
		if kind == "" {
			continue
		}
		// leaf if no seen child in tree
		isLeaf := true
		for _, cpid := range children[n.PID] {
			if seen[cpid] {
				isLeaf = false
				break
			}
		}
		cands = append(cands, cand{
			pid:   n.PID,
			cmd:   n.Cmd,
			kind:  kind,
			depth: depthOf[n.PID],
			leaf:  isLeaf,
		})
	}
	sort.SliceStable(cands, func(i, j int) bool {
		// Prefer deeper first; among equal depth prefer leaves.
		if cands[i].depth != cands[j].depth {
			return cands[i].depth > cands[j].depth
		}
		if cands[i].leaf != cands[j].leaf {
			return cands[i].leaf
		}
		return cands[i].pid < cands[j].pid
	})

	lsof := opts.Lsof
	if lsof == nil {
		lsof = func(int) []string { return nil }
	}

	result := &Result{
		InputPID: pid,
		Kind:     "none",
		Tree:     tree,
	}

	for _, c := range cands {
		files := lsof(c.pid)
		for _, f := range files {
			kind, sessionID, hit := parseSessionFromPath(f)
			if !hit {
				continue
			}
			// Prefer kind from path; fall back to candidate classification.
			if kind == "" {
				kind = c.kind
			}
			result.Kind = kind
			result.SessionID = sessionID
			result.Confidence = "hard"
			result.RunnerPID = c.pid
			result.RunnerCmd = c.cmd
			if c.pid == pid {
				result.Source = "open-files"
			} else {
				result.Source = "open-files+tree"
			}
			enrichResult(result, opts)
			return result, nil
		}
	}

	// Soft miss: process known, no session.
	_ = input
	return result, nil
}

// enrichResult fills GrokTitle/GrokModel when EnrichInfo is set and Kind is grok.
// Lookup errors are soft (warning only); they do not fail ResolveFromPID.
func enrichResult(result *Result, opts Options) {
	if result == nil || !opts.EnrichInfo {
		return
	}
	if result.Kind != "grok" || result.SessionID == "" {
		return
	}
	if opts.LookupGrokInfo == nil {
		return
	}
	title, model, err := opts.LookupGrokInfo(opts.GrokHome, result.SessionID)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("LookupGrokInfo: %v", err))
		return
	}
	result.GrokTitle = title
	result.GrokModel = model
}
