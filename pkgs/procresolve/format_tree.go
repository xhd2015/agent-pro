package procresolve

import (
	"fmt"
	"sort"
	"strings"
)

// FormatTree rebuilds a parent→child forest from node PPIDs and prints each
// node as connectors + "PID Cmd". Pure: does not call Resolve or Lsof.
//
// Connector rules (matches locked P2 templates for sole-child chains):
//   - Non-last siblings use branch (├── / +--) and pipe continuation (│ / |).
//   - Last siblings that still have children also use branch + pipe (so a
//     single intermediate child is drawn with ├──/│, not └──/spaces).
//   - Last siblings that are leaves use last (└── / `--) and blank continuation.
func FormatTree(nodes []ProcNode, opts TreeFormatOptions) string {
	if len(nodes) == 0 {
		return ""
	}

	byPID := make(map[int]ProcNode, len(nodes))
	inSet := make(map[int]bool, len(nodes))
	children := make(map[int][]int, len(nodes))
	for _, n := range nodes {
		byPID[n.PID] = n
		inSet[n.PID] = true
	}
	for _, n := range nodes {
		if inSet[n.PPID] {
			children[n.PPID] = append(children[n.PPID], n.PID)
		}
	}
	for pid := range children {
		sort.Ints(children[pid])
	}

	// Roots: PPID not in the node set.
	var roots []int
	for _, n := range nodes {
		if !inSet[n.PPID] {
			roots = append(roots, n.PID)
		}
	}
	sort.Ints(roots)
	// Fallback: closed graph / cycle — pick lowest PID as sole root.
	if len(roots) == 0 {
		for _, n := range nodes {
			roots = append(roots, n.PID)
		}
		sort.Ints(roots)
		if len(roots) > 0 {
			roots = roots[:1]
		}
	}

	branch, last, pipe, blank := treeConnectors(opts.ASCII)

	var b strings.Builder
	var walk func(pid int, linePrefix, childPrefix string)
	walk = func(pid int, linePrefix, childPrefix string) {
		n := byPID[pid]
		fmt.Fprintf(&b, "%s%d %s\n", linePrefix, n.PID, n.Cmd)

		kids := children[pid]
		for i, cpid := range kids {
			hasKids := len(children[cpid]) > 0
			isLast := i == len(kids)-1
			// Last leaf → last connector; otherwise branch (incl. last-with-kids).
			useLast := isLast && !hasKids
			var conn, cont string
			if useLast {
				conn, cont = last, blank
			} else {
				conn, cont = branch, pipe
			}
			walk(cpid, childPrefix+conn, childPrefix+cont)
		}
	}

	for _, r := range roots {
		walk(r, "", "")
	}
	return b.String()
}

func treeConnectors(ascii bool) (branch, last, pipe, blank string) {
	if ascii {
		// +--  / `--  / |   / four spaces — each connector slot is width 4.
		return "+-- ", "`-- ", "|   ", "    "
	}
	return "├── ", "└── ", "│   ", "    "
}
