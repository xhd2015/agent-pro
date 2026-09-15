package usage

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	codexmodels "github.com/xhd2015/agent-pro/agent/codex/models"
	"github.com/xhd2015/agent-pro/agent/commandcode"
	grokmodels "github.com/xhd2015/agent-pro/agent/grok/models"
	codexapi "github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/api"
	codexusage "github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/usage"
	grokapi "github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/api"
	grokusage "github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/usage"
)

// EnvFixtureDir points grok and codex fetches at JSON fixtures instead of
// their HTTP endpoints. Tests only; unset in normal runs.
const EnvFixtureDir = "AGENT_PRO_USAGE_FIXTURE_DIR"

// DefaultTimeout bounds one provider's usage fetch.
const DefaultTimeout = 30 * time.Second

// AllProviders is the provider set every collect covers, in output order.
func AllProviders() []ProviderID {
	return []ProviderID{Grok, Codex, CommandCode}
}

// CollectOptions configures Collect. Home fields empty mean the provider
// default ($GROK_HOME or ~/.grok, $CODEX_HOME or ~/.codex, ~/.commandcode).
type CollectOptions struct {
	// Now is the clock; nil means time.Now.
	Now func() time.Time
	// Timeout bounds each provider's usage fetch; 0 means DefaultTimeout.
	Timeout time.Duration
	// GrokHome is the grok home directory, used for both auth and sessions.
	GrokHome string
	// CodexHome is the codex home directory, used for both auth and sessions.
	CodexHome string
	// CommandCodeHome is the Command Code home directory.
	CommandCodeHome string
	// CommandCodeAPIURL overrides the Command Code API base URL; empty uses
	// COMMANDCODE_API_URL when COMMANDCODE_SANDBOX=true, else the default.
	CommandCodeAPIURL string
	// FixtureDir serves grok and codex usage from JSON files in this directory
	// instead of HTTP. Tests only.
	FixtureDir string
}

// Collect fetches every provider's usage over its API and counts each
// provider's on-disk sessions, in parallel, returning one Record per provider
// in AllProviders order. A provider whose usage fetch fails still yields a
// Record: Usage.OK is false with the error, and Sessions is unaffected because
// it only reads local files.
func Collect(ctx context.Context, opts CollectOptions) []Record {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}

	records := make([]Record, len(AllProviders()))
	var wg sync.WaitGroup
	for i, id := range AllProviders() {
		wg.Add(1)
		go func(i int, id ProviderID) {
			defer wg.Done()
			records[i] = collectProvider(ctx, id, opts)
		}(i, id)
	}
	wg.Wait()
	return records
}

// collectProvider builds one provider's record. Local session counting runs
// regardless of the usage outcome.
func collectProvider(ctx context.Context, id ProviderID, opts CollectOptions) Record {
	now := opts.Now().UTC().Truncate(time.Second)
	rec := Record{
		Schema:   SchemaID,
		TS:       now,
		Provider: id,
	}

	fetchCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	fetched := make(chan UsageBlock, 1)
	go func() {
		fetched <- fetchUsage(fetchCtx, id, opts)
	}()
	rec.Usage = <-fetched
	cancel()

	rec.Sessions = countSessions(id, opts)
	return rec
}

func fetchUsage(ctx context.Context, id ProviderID, opts CollectOptions) UsageBlock {
	start := time.Now()
	var (
		block UsageBlock
		err   error
	)
	switch id {
	case Grok:
		block, err = fetchGrokUsage(ctx, opts)
	case Codex:
		block, err = fetchCodexUsage(ctx, opts)
	case CommandCode:
		block, err = fetchCommandCodeUsage(ctx, opts)
	default:
		err = fmt.Errorf("unknown usage provider %q", id)
	}
	block.DurationMS = time.Since(start).Milliseconds()
	if err != nil {
		block.OK = false
		block.Error = err.Error()
	}
	return block
}

func fetchGrokUsage(ctx context.Context, opts CollectOptions) (UsageBlock, error) {
	authPath, err := providerAuthPath(opts.GrokHome, grokmodels.DefaultHome())
	if err != nil {
		return UsageBlock{}, err
	}
	fetchOpts := grokusage.FetchOpts{AuthPath: authPath}
	applyGrokFixture(&fetchOpts, opts.FixtureDir)

	snap, err := grokusage.Fetch(ctx, fetchOpts)
	if err != nil {
		return UsageBlock{}, err
	}

	block := UsageBlock{
		OK:       true,
		Endpoint: snap.Source,
		Values:   map[string]float64{},
		Display:  map[string]string{},
	}
	setInt64(block.Values, "used", snap.Used, snap.Used != 0)
	setInt64(block.Values, "monthly_limit", snap.MonthlyLimit, snap.MonthlyLimit != 0)
	setInt(block.Values, "used_percent", snap.UsedPercent, snap.UsedPercent >= 0)
	setInt(block.Values, "remaining_percent", snap.RemainingPercent, snap.RemainingPercent >= 0)
	setTime(block.Values, "period_start", snap.PeriodStart)
	setTime(block.Values, "period_end", snap.PeriodEnd)
	setTime(block.Values, "reset_at", snap.ResetAt)

	if snap.Email != "" {
		block.Display["email"] = snap.Email
	}
	if snap.PeriodType != "" {
		block.Display["period"] = snap.PeriodType
	}
	if !snap.ResetAt.IsZero() {
		block.Display["reset"] = snap.ResetAt.UTC().Format(time.RFC3339)
	}
	block.Display["usage"] = grokHeadline(snap)
	block.Display["detail"] = grokDetail(snap)
	return block, nil
}

// grokHeadline prefers the percentage the API reported, falling back to the
// raw used/limit pair when only those are present.
func grokHeadline(snap grokusage.Snapshot) string {
	period := snap.PeriodType
	var parts []string
	switch {
	case snap.UsedPercent >= 0:
		parts = append(parts, fmt.Sprintf("%d%% used", snap.UsedPercent))
	case snap.Used > 0 || snap.MonthlyLimit > 0:
		parts = append(parts, fmt.Sprintf("%d/%d used", snap.Used, snap.MonthlyLimit))
	}
	if period != "" {
		parts = append(parts, period)
	}
	if len(parts) == 0 {
		return "usage reported"
	}
	return strings.Join(parts, " · ")
}

func grokDetail(snap grokusage.Snapshot) string {
	var parts []string
	if snap.RemainingPercent >= 0 {
		parts = append(parts, fmt.Sprintf("%d%% remaining", snap.RemainingPercent))
	}
	if !snap.ResetAt.IsZero() {
		parts = append(parts, "resets "+snap.ResetAt.UTC().Format(time.RFC3339))
	}
	if snap.Email != "" {
		parts = append(parts, snap.Email)
	}
	return strings.Join(parts, " · ")
}

func fetchCodexUsage(ctx context.Context, opts CollectOptions) (UsageBlock, error) {
	authPath, err := providerAuthPath(opts.CodexHome, codexmodels.DefaultHome())
	if err != nil {
		return UsageBlock{}, err
	}
	fetchOpts := codexusage.FetchOpts{AuthPath: authPath}
	applyCodexFixture(&fetchOpts, opts.FixtureDir)

	snap, err := codexusage.Fetch(ctx, fetchOpts)
	if err != nil {
		return UsageBlock{}, err
	}

	block := UsageBlock{
		OK:       true,
		Endpoint: snap.Source,
		Values:   map[string]float64{},
		Display:  map[string]string{},
	}
	setInt(block.Values, "used_percent", snap.UsedPercent, snap.UsedPercent >= 0)
	setInt(block.Values, "remaining_percent", snap.RemainingPercent, snap.RemainingPercent >= 0)
	setTime(block.Values, "reset_at", snap.ResetAt)

	if snap.PlanType != "" {
		block.Display["plan_type"] = snap.PlanType
	}
	if snap.Email != "" {
		block.Display["email"] = snap.Email
	}
	if !snap.ResetAt.IsZero() {
		block.Display["reset"] = snap.ResetAt.UTC().Format(time.RFC3339)
	}
	block.Display["usage"] = codexHeadline(snap)
	block.Display["detail"] = codexDetail(snap)
	return block, nil
}

func codexHeadline(snap codexusage.Snapshot) string {
	switch {
	case snap.UsedPercent >= 0:
		return fmt.Sprintf("%d%% used", snap.UsedPercent)
	case snap.RemainingPercent >= 0:
		return fmt.Sprintf("%d%% remaining", snap.RemainingPercent)
	default:
		return "usage reported"
	}
}

func codexDetail(snap codexusage.Snapshot) string {
	var parts []string
	if snap.RemainingPercent >= 0 {
		parts = append(parts, fmt.Sprintf("%d%% remaining", snap.RemainingPercent))
	}
	if snap.PlanType != "" {
		parts = append(parts, snap.PlanType)
	}
	if !snap.ResetAt.IsZero() {
		parts = append(parts, "resets "+snap.ResetAt.UTC().Format(time.RFC3339))
	}
	if snap.Email != "" {
		parts = append(parts, snap.Email)
	}
	return strings.Join(parts, " · ")
}

// EndpointCommandCodeUsage labels the composite Command Code usage fetch.
const EndpointCommandCodeUsage = "/alpha/usage/summary"

func fetchCommandCodeUsage(ctx context.Context, opts CollectOptions) (UsageBlock, error) {
	home := strings.TrimSpace(opts.CommandCodeHome)
	if home == "" {
		home = commandcode.DefaultHome()
	}
	client, err := commandcode.NewClient(home, opts.CommandCodeAPIURL)
	if err != nil {
		return UsageBlock{}, err
	}
	data, err := client.FetchUsage(ctx)
	if err != nil {
		return UsageBlock{}, err
	}
	// No identity means no scoped endpoint was queried at all.
	if data.Whoami == nil {
		if len(data.Errors) > 0 {
			return UsageBlock{}, fmt.Errorf("%s", data.Errors[0])
		}
		return UsageBlock{}, fmt.Errorf("no account data returned")
	}

	view := commandcode.ProjectUsageView(data, opts.now())
	block := UsageBlock{
		OK:       true,
		Endpoint: EndpointCommandCodeUsage,
		Values:   map[string]float64{},
		Display:  map[string]string{},
	}
	if view.Credits.HasCreditsInfo {
		block.Values["usage_percent"] = round2(view.Credits.UsagePercent)
		block.Values["credits_remaining"] = round2(view.Credits.TotalRemaining)
		block.Values["cost_usd"] = round2(view.Credits.TotalSpent)
	}
	if view.WindowLimits != nil {
		if view.WindowLimits.FiveHour != nil && view.WindowLimits.FiveHour.Cap > 0 {
			block.Values["five_hour_used_percent"] = windowPercent(view.WindowLimits.FiveHour)
		}
		if view.WindowLimits.Weekly != nil && view.WindowLimits.Weekly.Cap > 0 {
			block.Values["weekly_used_percent"] = windowPercent(view.WindowLimits.Weekly)
			if view.WindowLimits.Weekly.ResetAt > 0 {
				block.Values["weekly_reset_at"] = view.WindowLimits.Weekly.ResetAt
			}
		}
	}
	if view.Summary != nil {
		block.Values["requests_total"] = float64(view.Summary.TotalCount)
		block.Values["requests_failed_total"] = float64(view.Summary.FailedCount)
		if view.Summary.TotalTokens > 0 {
			block.Values["tokens_in_total"] = float64(view.Summary.TotalTokensIn)
			block.Values["tokens_out_total"] = float64(view.Summary.TotalTokensOut)
		}
	}
	if view.DaysLeft != nil {
		block.Values["days_to_renewal"] = float64(*view.DaysLeft)
	}
	if view.Plan != nil {
		block.Display["plan"] = view.Plan.Name
	}
	if data.Whoami.User.Email != "" {
		block.Display["email"] = data.Whoami.User.Email
	}
	if view.UsageURL != "" {
		block.Display["usage_url"] = view.UsageURL
	}
	if data.Subscription != nil && data.Subscription.Status != "" {
		block.Display["subscription_status"] = data.Subscription.Status
	}
	block.Display["usage"] = commandCodeHeadline(view)
	block.Display["detail"] = commandCodeDetail(view)
	return block, nil
}

func commandCodeHeadline(view *commandcode.View) string {
	var parts []string
	if view.Credits.HasCreditsInfo {
		parts = append(parts, fmt.Sprintf("%.0f%% used", view.Credits.UsagePercent))
	}
	if view.Plan != nil {
		parts = append(parts, view.Plan.Name)
	}
	if len(parts) == 0 {
		return "usage reported"
	}
	return strings.Join(parts, " · ")
}

func commandCodeDetail(view *commandcode.View) string {
	var parts []string
	if view.Credits.HasCreditsInfo {
		parts = append(parts, fmt.Sprintf("$%.2f left", view.Credits.TotalRemaining))
	}
	if view.Summary != nil {
		parts = append(parts, fmt.Sprintf("%d requests", view.Summary.TotalCount))
	}
	if view.DaysLeft != nil {
		parts = append(parts, fmt.Sprintf("%d days to renewal", *view.DaysLeft))
	}
	return strings.Join(parts, " · ")
}

// windowPercent is the share of a rate-limit window already consumed.
func windowPercent(span *commandcode.WindowSpan) float64 {
	if span == nil || span.Cap <= 0 {
		return 0
	}
	return round2(span.Used / span.Cap * 100)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func (opts CollectOptions) now() time.Time {
	if opts.Now == nil {
		return time.Now()
	}
	return opts.Now()
}

// providerAuthPath resolves the auth.json path for a provider home.
func providerAuthPath(home, fallback string) (string, error) {
	dir := strings.TrimSpace(home)
	if dir == "" {
		dir = fallback
	}
	if dir == "" {
		return "", fmt.Errorf("cannot determine the provider home directory")
	}
	return filepath.Join(dir, "auth.json"), nil
}

// applyGrokFixture swaps the grok HTTP transport for fixture files when a
// fixture directory is configured.
func applyGrokFixture(opts *grokusage.FetchOpts, fixtureDir string) {
	dir := strings.TrimSpace(fixtureDir)
	if dir == "" {
		return
	}
	opts.GetJSON = func(_ context.Context, getOpts grokapi.GetOpts) ([]byte, error) {
		return readFixtureFile(dir, grokFixtureNames[getOpts.URL])
	}
	// Fixture tokens never expire, so skip the token refresh round trip.
	opts.Ensure = func(_ context.Context, ensureOpts grokapi.EnsureOpts) (grokapi.Auth, error) {
		return ensureOpts.Auth, nil
	}
}

// applyCodexFixture swaps the codex HTTP transport for fixture files.
func applyCodexFixture(opts *codexusage.FetchOpts, fixtureDir string) {
	dir := strings.TrimSpace(fixtureDir)
	if dir == "" {
		return
	}
	opts.GetJSON = func(_ context.Context, getOpts codexapi.GetOpts) ([]byte, error) {
		return readFixtureFile(dir, codexFixtureNames[getOpts.URL])
	}
}

// readFixtureFile reads one fixture body. An unmapped URL or a missing file is
// an error, which lets leaves exercise the codex fallback and failure paths.
func readFixtureFile(dir, name string) ([]byte, error) {
	if name == "" {
		return nil, fmt.Errorf("no fixture configured")
	}
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return nil, fmt.Errorf("fixture %s: %w", name, err)
	}
	return raw, nil
}

// grokFixtureNames maps grok API URLs to fixture file names.
var grokFixtureNames = map[string]string{
	grokapi.BillingURL:        "grok-billing.json",
	grokapi.BillingCreditsURL: "grok-billing-credits.json",
}

// codexFixtureNames maps codex API URLs to fixture file names.
var codexFixtureNames = map[string]string{
	codexapi.CodexUsageURL: "codex-usage.json",
	codexapi.WhamUsageURL:  "codex-wham.json",
}

func setTime(values map[string]float64, key string, at time.Time) {
	if at.IsZero() {
		return
	}
	values[key] = float64(at.Unix())
}

func setInt(values map[string]float64, key string, value int, ok bool) {
	if !ok {
		return
	}
	values[key] = float64(value)
}

func setInt64(values map[string]float64, key string, value int64, ok bool) {
	if !ok {
		return
	}
	values[key] = float64(value)
}
