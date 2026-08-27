package agentruncli

import (
	"time"

	"github.com/xhd2015/agent-pro/pkgs/tty/detection/idle"
)

const defaultIdleGrace = idle.DefaultGrace

// idleWatchFirstDelay / idleWatchSamplesPerCycle kept for older call sites.
const idleWatchFirstDelay = idle.FirstDelay
const idleWatchSamplesPerCycle = idle.SamplesPerCycle

// IdleWatchSchedule is the serve-loop sleep plan (delegates to idle.Schedule).
func IdleWatchSchedule(timeout time.Duration) (first, gap time.Duration) {
	return idle.Schedule(timeout)
}
