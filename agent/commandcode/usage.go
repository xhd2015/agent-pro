package commandcode

import (
	"context"
	"fmt"
	"sync"
)

// Send-error strings the cmd CLI surfaces in its usage overlay.
const (
	ErrNotAuthenticated = "Not authenticated"
	ErrSessionExpired   = "Session expired"
	ErrNetwork          = "Network error: unable to reach API"
)

// UsageData is the merged result of the composite usage fetch. A nil field
// means that endpoint failed; Errors records one message per failure. The JSON
// tags match the shape the cmd CLI's fetchUsageData returns.
type UsageData struct {
	Whoami       *Whoami       `json:"whoami"`
	Credits      *Credits      `json:"credits"`
	Subscription *Subscription `json:"subscription"`
	Summary      *UsageSummary `json:"summary"`
	Errors       []string      `json:"errors"`
}

// Whoami fetches account identity. orgID is optional and only meaningful for
// team namespaces.
func (c *Client) Whoami(ctx context.Context, orgID string) (*Whoami, error) {
	var out Whoami
	params := map[string]string{"limits": "1", "orgId": orgID}
	if err := c.Get(ctx, PathWhoami, params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Credits fetches the credit balance and rate-limit windows.
func (c *Client) Credits(ctx context.Context, orgID string) (*Credits, error) {
	var out Credits
	if err := c.Get(ctx, PathCredits, map[string]string{"orgId": orgID}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Subscription fetches the current subscription, or nil when the account has
// none.
func (c *Client) Subscription(ctx context.Context, orgID string) (*Subscription, error) {
	var out SubscriptionResponse
	if err := c.Get(ctx, PathSubscriptions, map[string]string{"orgId": orgID}, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// Summary fetches usage totals, bounded by since (an RFC3339 timestamp) when
// set. An empty since reports the whole account history.
func (c *Client) Summary(ctx context.Context, orgID, since string) (*UsageSummary, error) {
	var out UsageSummary
	if err := c.Get(ctx, PathUsageSummary, map[string]string{"orgId": orgID, "since": since}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Namespaces fetches the personal and team namespaces.
func (c *Client) Namespaces(ctx context.Context) (*Namespaces, error) {
	var out Namespaces
	if err := c.Get(ctx, PathNamespaces, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UsageOptions overrides parts of the composite fetch.
type UsageOptions struct {
	// OrgID pins the org scope. Empty uses whoami's org, else the personal
	// namespace.
	OrgID string
	// Since pins the summary period start (RFC3339). Empty uses the
	// subscription's current period start.
	Since string
}

// FetchUsage mirrors the cmd CLI's usage overlay fetch: whoami supplies the
// org scope, credits and subscription load concurrently, and the subscription
// period start bounds the summary. Per-endpoint failures are collected in
// UsageData.Errors instead of aborting, so a partial view still renders.
func (c *Client) FetchUsage(ctx context.Context) (*UsageData, error) {
	return c.FetchUsageWithOptions(ctx, UsageOptions{})
}

// FetchUsageWithOptions is FetchUsage with explicit org and period overrides.
func (c *Client) FetchUsageWithOptions(ctx context.Context, opts UsageOptions) (*UsageData, error) {
	out := &UsageData{Errors: []string{}}

	whoami, err := c.Whoami(ctx, "")
	if err != nil {
		out.Errors = append(out.Errors, mapUsageError(PathWhoami, err))
		return out, nil
	}
	out.Whoami = whoami

	orgID := opts.OrgID
	if orgID == "" && whoami.Org != nil {
		orgID = whoami.Org.ID
	}

	var (
		wg         sync.WaitGroup
		creditsErr string
		subErr     string
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		credits, err := c.Credits(ctx, orgID)
		if err != nil {
			creditsErr = mapUsageError(PathCredits, err)
			return
		}
		out.Credits = credits
	}()
	go func() {
		defer wg.Done()
		sub, err := c.Subscription(ctx, orgID)
		if err != nil {
			subErr = mapUsageError(PathSubscriptions, err)
			return
		}
		out.Subscription = sub
	}()
	wg.Wait()
	out.Errors = appendNonEmpty(out.Errors, creditsErr, subErr)

	var since string
	if opts.Since != "" {
		since = opts.Since
	} else if out.Subscription != nil {
		since = out.Subscription.CurrentPeriodStart
	}
	summary, err := c.Summary(ctx, orgID, since)
	if err != nil {
		out.Errors = appendNonEmpty(out.Errors, mapUsageError(PathUsageSummary, err))
		return out, nil
	}
	out.Summary = summary
	return out, nil
}

// mapUsageError converts a transport error into the CLI's overlay wording.
func mapUsageError(endpoint string, err error) string {
	switch typed := err.(type) {
	case *NetworkError:
		return ErrNetwork
	case *APIError:
		if typed.Unauthorized() {
			return ErrSessionExpired
		}
		return fmt.Sprintf("%s: %s", endpoint, typed.describe())
	default:
		return fmt.Sprintf("%s: %v", endpoint, err)
	}
}

func appendNonEmpty(errs []string, msgs ...string) []string {
	for _, msg := range msgs {
		if msg != "" {
			errs = append(errs, msg)
		}
	}
	return errs
}
