package usage

import (
	"context"
	"fmt"

	codextty "github.com/xhd2015/agent-pro/agent/codex/tty"
	"github.com/xhd2015/agent-pro/agent/debuglog"
	groktty "github.com/xhd2015/agent-pro/agent/grok/tty"
)

// ProviderID identifies a usage fetch provider.
type ProviderID string

const (
	Grok  ProviderID = "grok"
	Codex ProviderID = "codex"
)

// Snapshot is the normalized usage payload returned by Fetch.
type Snapshot struct {
	Provider     ProviderID
	UsagePercent string
	Reset        string
	CreditsUsed  string
	CreditsTotal string
}

// Fetch retrieves usage for the given provider in-process.
func Fetch(ctx context.Context, id ProviderID) (*Snapshot, error) {
	debuglog.Log(debuglog.Entry{
		Event: "fetch_dispatch",
		Labels: map[string]string{
			"component": "usage",
			"provider":  string(id),
			"phase":     "fetch",
		},
	})

	switch id {
	case Grok:
		info, err := groktty.FetchUsage(ctx)
		if err != nil {
			logUsageError(Grok, err)
			return nil, err
		}
		snap := &Snapshot{
			Provider:     Grok,
			UsagePercent: info.WeeklyLimit,
			Reset:        info.NextReset,
		}
		logUsageDone(snap)
		return snap, nil
	case Codex:
		info, err := codextty.FetchStatus(ctx)
		if err != nil {
			logUsageError(Codex, err)
			return nil, err
		}
		snap := &Snapshot{
			Provider:     Codex,
			UsagePercent: info.MonthlyUsage,
			Reset:        info.NextReset,
			CreditsUsed:  info.CreditsUsed,
			CreditsTotal: info.CreditsTotal,
		}
		logUsageDone(snap)
		return snap, nil
	default:
		err := fmt.Errorf("unknown usage provider %q", id)
		logUsageError(id, err)
		return nil, err
	}
}

func logUsageError(id ProviderID, err error) {
	debuglog.Log(debuglog.Entry{
		Event: "fetch_error",
		Labels: map[string]string{
			"component": "usage",
			"provider":  string(id),
			"phase":     "error",
		},
		Fields: map[string]any{"error": err.Error()},
	})
}

func logUsageDone(snap *Snapshot) {
	if snap == nil {
		return
	}
	debuglog.Log(debuglog.Entry{
		Event: "fetch_done",
		Labels: map[string]string{
			"component": "usage",
			"provider":  string(snap.Provider),
			"phase":     "fetch",
		},
		Fields: map[string]any{
			"usage_percent": snap.UsagePercent,
			"reset":         snap.Reset,
			"credits_used":  snap.CreditsUsed,
			"credits_total": snap.CreditsTotal,
		},
	})
}