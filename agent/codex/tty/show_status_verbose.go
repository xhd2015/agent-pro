package tty

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// verboseLog emits -v / -debug diagnostics to stderr.
type verboseLog struct {
	enabled bool
	start   time.Time
	last    time.Time
}

func newVerboseLog(enabled bool) *verboseLog {
	now := time.Now()
	return &verboseLog{enabled: enabled, start: now, last: now}
}

func (v *verboseLog) enabledf() bool {
	return v != nil && v.enabled
}

func (v *verboseLog) timing() (total, step time.Duration) {
	now := time.Now()
	total = now.Sub(v.start)
	step = now.Sub(v.last)
	v.last = now
	return total, step
}

func (v *verboseLog) printf(format string, args ...any) {
	if !v.enabledf() {
		return
	}
	fmt.Fprintf(os.Stderr, "codex-show-status: "+format+"\n", args...)
}

func (v *verboseLog) command(argv []string) {
	if !v.enabledf() {
		return
	}
	total, step := v.timing()
	v.printf("[%s +%s] command: %s", fmtDuration(total), fmtDuration(step), strings.Join(argv, " "))
}

func (v *verboseLog) stateChange(label, detail string) {
	if !v.enabledf() {
		return
	}
	total, step := v.timing()
	if detail == "" {
		v.printf("[%s +%s] %s", fmtDuration(total), fmtDuration(step), label)
		return
	}
	v.printf("[%s +%s] %s (%s)", fmtDuration(total), fmtDuration(step), label, detail)
}

func (v *verboseLog) snapshot(phase string, d time.Duration, detail string) {
	if !v.enabledf() {
		return
	}
	total, step := v.timing()
	if detail == "" {
		v.printf("[%s +%s] snapshot %s took %s", fmtDuration(total), fmtDuration(step), phase, fmtDuration(d))
		return
	}
	v.printf("[%s +%s] snapshot %s took %s (%s)", fmtDuration(total), fmtDuration(step), phase, fmtDuration(d), detail)
}

func (v *verboseLog) phaseDone(phase string, since time.Time) {
	if !v.enabledf() {
		return
	}
	total, step := v.timing()
	v.printf("[%s +%s] phase %s done (phase %s)", fmtDuration(total), fmtDuration(step), phase, fmtDuration(time.Since(since)))
}

func fmtDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	ms := d.Round(time.Millisecond)
	if ms < time.Second {
		return fmt.Sprintf("%dms", ms.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", ms.Seconds())
}