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
}

func newVerboseLog(enabled bool) *verboseLog {
	return &verboseLog{enabled: enabled, start: time.Now()}
}

func (v *verboseLog) enabledf() bool {
	return v != nil && v.enabled
}

func (v *verboseLog) printf(format string, args ...any) {
	if !v.enabledf() {
		return
	}
	fmt.Fprintf(os.Stderr, "grok-show-usage: "+format+"\n", args...)
}

func (v *verboseLog) command(argv []string) {
	if !v.enabledf() {
		return
	}
	v.printf("command: %s", strings.Join(argv, " "))
}

func (v *verboseLog) attempt(n int) {
	v.printf("attempt %d", n)
}

func (v *verboseLog) stateChange(label, detail string) {
	if !v.enabledf() {
		return
	}
	elapsed := time.Since(v.start).Round(100 * time.Millisecond)
	if detail == "" {
		v.printf("[%s] state: %s", elapsed, label)
		return
	}
	v.printf("[%s] state: %s (%s)", elapsed, label, detail)
}