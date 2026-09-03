//go:build !darwin

package sessions

import "context"

// waitProcessesExit is a no-op on non-Darwin; readiness falls back to flock.
func waitProcessesExit(ctx context.Context, pids []int) error {
	_ = pids
	<-ctx.Done()
	return ctx.Err()
}
