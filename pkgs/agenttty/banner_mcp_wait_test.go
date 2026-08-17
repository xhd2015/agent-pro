package agenttty

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForBannerRemoteOpts_mcpStartingDoesNotConsumeTimeout(t *testing.T) {
	orig := bannerSnapshotText
	t.Cleanup(func() { bannerSnapshotText = orig })

	const mcp = "OpenAI Codex\n• Starting MCP servers (0/8): slowinit_01\n› prompt\n"
	const idle = "OpenAI Codex\nCodex › ready\n"
	var n atomic.Int32
	bannerSnapshotText = func(_, _ string) (string, error) {
		if n.Add(1) <= 8 {
			return mcp, nil
		}
		return idle, nil
	}

	// 150ms budget would expire if MCP polls counted (~8*75ms).
	start := time.Now()
	err := waitForBannerRemoteOpts(context.Background(), "127.0.0.1:1", "s", "codex", nil, 150*time.Millisecond, nil)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("want success after long MCP then idle: %v (polls=%d elapsed=%v)", err, n.Load(), elapsed)
	}
	if elapsed < 400*time.Millisecond {
		t.Fatalf("expected MCP phase to hold the waiter, elapsed=%v polls=%d", elapsed, n.Load())
	}
}

func TestWaitForBannerRemoteOpts_nonMCPStillTimesOut(t *testing.T) {
	orig := bannerSnapshotText
	t.Cleanup(func() { bannerSnapshotText = orig })
	bannerSnapshotText = func(_, _ string) (string, error) {
		return "still loading chrome with no prompt", nil
	}
	start := time.Now()
	err := waitForBannerRemoteOpts(context.Background(), "127.0.0.1:1", "s", "codex", nil, 120*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected banner timeout when MCP is not starting")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("non-MCP timeout should fire on the short budget, took %v", time.Since(start))
	}
}

func TestBannerDetected_mcpStartingFalse_idleTrue(t *testing.T) {
	mcp := []byte("OpenAI Codex\n• Starting MCP servers (0/8): slowinit_01\n› prompt\n")
	if BannerDetected(mcp, "codex", []string{"CODEX_TTY_BANNER"}) {
		t.Fatal("BannerDetected must be false during Starting MCP servers")
	}
	idle := []byte("OpenAI Codex\nCodex › ready\n")
	if !BannerDetected(idle, "codex", []string{"CODEX_TTY_BANNER"}) {
		t.Fatal("BannerDetected must be true on idle Codex ›")
	}
	incomplete := []byte("OpenAI Codex\nMCP startup incomplete (failed: computer-use)\n› Summarize recent commits\n")
	if !BannerDetected(incomplete, "codex", []string{"CODEX_TTY_BANNER"}) {
		t.Fatal("MCP startup incomplete + › must still be inject-ready")
	}
}
