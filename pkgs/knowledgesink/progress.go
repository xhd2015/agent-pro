package knowledgesink

import (
	"fmt"
	"io"
	"strings"
)

// Stage kinds for default progress (spl codelens dev–style [n/total] lines).
const (
	stageLaunch   = "launch"
	stageAgent    = "agent"
	stageValidate = "validate"
	stageShip     = "ship"
	stageMerge    = "merge"
)

func progressWriter(opts Opts) io.Writer {
	return opts.Stderr
}

// progress prints "[n/total] kind         msg" like spl codelens dev.
func progress(opts Opts, n, total int, kind, msg string) {
	w := progressWriter(opts)
	if w == nil || total <= 0 {
		return
	}
	fmt.Fprintf(w, "[%d/%d] %-12s %s\n", n, total, kind, msg)
}

func stageTotal(opts Opts) int {
	if opts.CreateMR {
		if opts.AutoMergeMR {
			return 5
		}
		return 4
	}
	return 2
}

func launchMsg(opts Opts, mode Mode) string {
	parts := []string{string(mode)}
	if mode == "" {
		parts[0] = string(ModeHeadless)
	}
	if opts.CreateMR {
		parts = append(parts, "create-mr")
	}
	if opts.AutoMergeMR {
		parts = append(parts, "auto-merge")
	}
	return strings.Join(parts, " ")
}

func writeDryRunStages(opts Opts, mode Mode) {
	total := stageTotal(opts)
	n := 0
	n++
	progress(opts, n, total, stageLaunch, "skip (dry-run) "+launchMsg(opts, mode))
	n++
	progress(opts, n, total, stageAgent, "skip (dry-run)")
	if opts.CreateMR {
		n++
		progress(opts, n, total, stageValidate, "skip (dry-run)")
		n++
		progress(opts, n, total, stageShip, "skip (dry-run)")
		if opts.AutoMergeMR {
			n++
			progress(opts, n, total, stageMerge, "skip (dry-run)")
		}
	}
}
