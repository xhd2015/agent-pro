package commandcode

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ANSI styles matching the cmd CLI's usage overlay color roles.
const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiCyan   = "\x1b[36m"
	ansiGray   = "\x1b[90m"
	ansiBadge  = "\x1b[1;44;97m" // bold, blue background, near-white text
)

// DefaultTerminalWidth is the fallback width for --width.
const DefaultTerminalWidth = 80

// CreditsView mirrors the credential projection the cmd CLI's overlay uses.
type CreditsView struct {
	MonthlyRemaining   float64
	PurchasedRemaining float64
	FreeRemaining      float64
	TotalRemaining     float64
	TotalSpent         float64
	TotalPool          float64
	UsagePercent       float64
	HasCreditsInfo     bool
}

// View is the projected usage view that FormatUsage renders.
type View struct {
	Subscription    *Subscription
	Plan            *Plan
	HasBillingData  bool
	UsageURL        string
	UsageURLDisplay string
	Credits         CreditsView
	WindowLimits    *WindowLimits
	OrgLimits       []OrgLimit
	Summary         *UsageSummary
	DaysLeft        *int
	Now             time.Time
}

// ProjectUsageView reproduces the cmd CLI's projectUsageView. now defaults to
// time.Now when zero.
func ProjectUsageView(data *UsageData, now time.Time) *View {
	if now.IsZero() {
		now = time.Now()
	}
	view := &View{Now: now}
	if data == nil {
		return view
	}

	view.Subscription = data.Subscription
	if data.Subscription != nil {
		view.Plan = GetPlanInfo(data.Subscription.PlanID)
	}
	if data.Credits != nil {
		view.WindowLimits = data.Credits.WindowLimits
	}
	if data.Whoami != nil {
		view.OrgLimits = data.Whoami.OrgLimits
	}
	view.Summary = data.Summary
	view.HasBillingData = data.Credits != nil || data.Subscription != nil

	var monthly, purchased, free float64
	if data.Credits != nil {
		balance := data.Credits.Credits
		monthly = math.Max(0, balance.MonthlyCredits)
		purchased = math.Max(0, balance.PurchasedCredits)
		free = math.Max(0, balance.FreeCredits)
	}
	totalRemaining := monthly + purchased + free

	var totalSpent float64
	if data.Summary != nil {
		totalSpent = math.Max(0, data.Summary.TotalCost)
	}

	// An active subscription measures usage against the plan allowance; other
	// statuses fall back to spend plus remaining balance.
	var planAllowance *float64
	if data.Subscription != nil && data.Subscription.Status == "active" && view.Plan != nil {
		allowance := view.Plan.MonthlyCredits
		planAllowance = &allowance
	}
	var totalPool float64
	if planAllowance != nil {
		totalPool = math.Max(*planAllowance, monthly) + purchased + free
	} else {
		totalPool = totalSpent + totalRemaining
	}

	hasCreditsInfo := totalRemaining > 0 || totalSpent > 0
	var usagePercent float64
	if hasCreditsInfo {
		usagePercent = usagePercentOf(totalPool-totalRemaining, totalPool)
	}

	if data.Whoami != nil {
		view.UsageURL = studioUsageURL(data.Whoami, DefaultStudioHost)
		if view.UsageURL != "" {
			view.UsageURLDisplay = stripScheme(view.UsageURL)
		}
	}
	if data.Subscription != nil && data.Subscription.CurrentPeriodEnd != "" {
		view.DaysLeft = daysRemaining(data.Subscription.CurrentPeriodEnd, now)
	}

	view.Credits = CreditsView{
		MonthlyRemaining:   monthly,
		PurchasedRemaining: purchased,
		FreeRemaining:      free,
		TotalRemaining:     totalRemaining,
		TotalSpent:         totalSpent,
		TotalPool:          totalPool,
		UsagePercent:       usagePercent,
		HasCreditsInfo:     hasCreditsInfo,
	}
	return view
}

// FormatOptions controls rendered width and color.
type FormatOptions struct {
	// Width is the terminal width in columns; 0 uses DefaultTerminalWidth.
	Width int
	// Color enables ANSI styling.
	Color bool
}

// FormatUsage renders the usage view with the same content, bars, colors, and
// spacing as the cmd CLI's usage overlay, minus its interactive footer.
func FormatUsage(view *View, opts FormatOptions) string {
	width := opts.Width
	if width <= 0 {
		width = DefaultTerminalWidth
	}
	barWidth := ProgressBarWidth(width)
	paint := painter{color: opts.Color}

	var b strings.Builder

	// Header: USAGE badge, plan name, status.
	b.WriteString(paint.paint(ansiBadge, " USAGE "))
	if view.Plan != nil {
		b.WriteString(" ")
		b.WriteString(paint.paint(ansiGray, view.Plan.Name+" Plan"))
	}
	if status := subscriptionStatus(view.Subscription); status != "" {
		b.WriteString(paint.paint(ansiDim, " · "))
		code := ansiYellow
		if status == "active" {
			code = ansiGreen
		}
		b.WriteString(paint.paint(code, status))
	}
	b.WriteString("\n\n")

	if !view.HasBillingData {
		b.WriteString("No billing data found.\n")
		if view.UsageURL != "" {
			b.WriteString(paint.paint(ansiDim, "Visit "))
			b.WriteString(paint.paint(ansiCyan, "Studio"))
			b.WriteString(paint.paint(ansiDim, " for usage details."))
			b.WriteString("\n")
		}
		writeOrgLimitsSection(&b, view.OrgLimits, barWidth, view.Now, paint)
		return b.String()
	}

	credits := view.Credits
	if credits.HasCreditsInfo {
		code := usageColorCode(credits.UsagePercent)
		bar := BuildBlockBar(credits.UsagePercent, barWidth)
		b.WriteString(paint.paint(code, bar.Filled))
		b.WriteString(paint.paint(ansiGray, bar.Empty))
		b.WriteString(paint.paint(ansiBold+code, fmt.Sprintf(" %d%% used", int(math.Round(credits.UsagePercent)))))
		b.WriteString("\n")
	} else {
		b.WriteString(paint.paint(ansiDim, "Plan details unavailable"))
		b.WriteString("\n")
	}

	if credits.HasCreditsInfo {
		if view.Subscription != nil && view.Subscription.CurrentPeriodEnd != "" {
			b.WriteString(paint.paint(ansiDim, "Cycle: "))
		}
		b.WriteString(FormatCredits(credits.TotalRemaining))
		b.WriteString(paint.paint(ansiDim, " left"))
		if view.Summary != nil {
			b.WriteString(paint.paint(ansiDim, " · "))
			b.WriteString(groupDigits(int64(view.Summary.TotalCount)))
			b.WriteString(paint.paint(ansiDim, " requests"))
		}
		if view.DaysLeft != nil {
			b.WriteString(paint.paint(ansiDim, " · "))
			code := daysColorCode(*view.DaysLeft)
			if *view.DaysLeft == 0 {
				b.WriteString(paint.paint(code, "renewal today"))
			} else {
				noun := "days"
				if *view.DaysLeft == 1 {
					noun = "day"
				}
				b.WriteString(paint.paint(code, fmt.Sprintf("%d %s", *view.DaysLeft, noun)))
				b.WriteString(paint.paint(ansiDim, " to renewal"))
			}
		}
		b.WriteString("\n")
	}

	if view.WindowLimits != nil && view.WindowLimits.Limited {
		b.WriteString("\n")
		b.WriteString(paint.paint(ansiDim, "Usage limits"))
		b.WriteString("\n")
		if window := view.WindowLimits.FiveHour; window != nil {
			b.WriteString(formatWindowLimitMeter("5-hour", window, barWidth, view.Now, paint))
			b.WriteString("\n")
		}
		if window := view.WindowLimits.Weekly; window != nil {
			if view.WindowLimits.FiveHour != nil {
				b.WriteString("\n")
			}
			b.WriteString(formatWindowLimitMeter("Weekly", window, barWidth, view.Now, paint))
			b.WriteString("\n")
		}
	}

	writeOrgLimitsSection(&b, view.OrgLimits, barWidth, view.Now, paint)

	if view.UsageURL != "" {
		b.WriteString("\n")
		b.WriteString(paint.paint(ansiDim, "Full breakdown at "))
		b.WriteString(paint.paint(ansiCyan, view.UsageURLDisplay))
		b.WriteString("\n")
	}
	return b.String()
}

// Bar is a rendered two-tone progress bar.
type Bar struct {
	Filled string
	Empty  string
}

// BuildBlockBar reproduces the cmd CLI's block bar, including its rounding
// guard so a nonzero percentage always shows one filled block and a sub-100
// percentage never fills the entire bar.
func BuildBlockBar(percentage float64, width int) Bar {
	if width < 0 {
		width = 0
	}
	clamped := math.Max(0, math.Min(100, percentage))
	filled := int(math.Round(clamped / 100 * float64(width)))
	if clamped > 0.5 && filled == 0 {
		filled = 1
	}
	if clamped < 99.5 && filled == width {
		filled = width - 1
	}
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return Bar{
		Filled: strings.Repeat("█", filled),
		Empty:  strings.Repeat("░", width-filled),
	}
}

// ProgressBarWidth reproduces the cmd CLI's bar width for a terminal width.
func ProgressBarWidth(terminalWidth int) int {
	switch {
	case terminalWidth < 50:
		return maxInt(10, terminalWidth-10)
	case terminalWidth < 70:
		return maxInt(15, terminalWidth-15)
	default:
		return minInt(30, terminalWidth-15)
	}
}

// FormatCredits renders a credit amount the way the CLI does ($X.XX).
func FormatCredits(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

// FormatUsageJSON renders the merged usage payload as indented JSON.
func FormatUsageJSON(data *UsageData) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func formatWindowLimitMeter(label string, window *WindowSpan, width int, now time.Time, paint painter) string {
	percent := 0.0
	if window.Cap > 0 {
		percent = math.Min(100, window.Used/window.Cap*100)
	}
	code := usageColorCode(percent)
	bar := BuildBlockBar(percent, width)

	var b strings.Builder
	b.WriteString(paint.paint(ansiDim, padRight(label, 7)+" "))
	b.WriteString(paint.paint(code, bar.Filled))
	b.WriteString(paint.paint(ansiGray, bar.Empty))
	b.WriteString(paint.paint(ansiBold+code, fmt.Sprintf(" %d%%", int(math.Round(percent)))))
	if resetAt := windowResetTime(window.ResetAt); resetAt.After(now) {
		b.WriteString(paint.paint(ansiDim, " · resets in "+FormatDuration(resetAt.Sub(now))))
	}
	return b.String()
}

func writeOrgLimitsSection(b *strings.Builder, limits []OrgLimit, width int, now time.Time, paint painter) {
	rows := projectOrgSpendLimitRows(limits, now)
	if len(rows) == 0 {
		return
	}
	b.WriteString("\n")
	b.WriteString(paint.paint(ansiDim, "Spend limits"))
	b.WriteString("\n")
	for i, row := range rows {
		if i > 0 {
			b.WriteString("\n")
		}
		code := usageColorCode(row.Percent)
		if row.Reached {
			code = ansiRed
		}
		bar := BuildBlockBar(row.Percent, width)
		b.WriteString(paint.paint(ansiDim, row.Label+" "))
		b.WriteString(paint.paint(code, bar.Filled))
		b.WriteString(paint.paint(ansiGray, bar.Empty))
		b.WriteString(paint.paint(ansiBold+code, fmt.Sprintf(" %d%%", int(math.Round(row.Percent)))))
		b.WriteString("\n")
		b.WriteString(paint.paint(ansiDim, strings.Repeat(" ", len(row.Label)+1)+row.Detail))
		b.WriteString("\n")
	}
}

// orgSpendLimitRow is one rendered spend-limit row.
type orgSpendLimitRow struct {
	Label   string
	Percent float64
	Detail  string
	Reached bool
}

func projectOrgSpendLimitRows(limits []OrgLimit, now time.Time) []orgSpendLimitRow {
	if len(limits) == 0 {
		return nil
	}
	labels := make([]string, len(limits))
	labelWidth := 7
	for i, limit := range limits {
		labels[i] = orgSpendLimitLabel(limit)
		if len(labels[i]) > labelWidth {
			labelWidth = len(labels[i])
		}
	}
	rows := make([]orgSpendLimitRow, len(limits))
	for i, limit := range limits {
		rows[i] = orgSpendLimitRow{
			Label:   padRight(labels[i], labelWidth),
			Percent: orgSpendLimitPercent(limit),
			Detail:  orgSpendLimitDetail(limit, now),
			Reached: orgSpendLimitReached(limit),
		}
	}
	return rows
}

const orgSpendLimitModelLabelMax = 18

func orgSpendLimitLabel(limit OrgLimit) string {
	if limit.Scope != "model" {
		return "Org-wide"
	}
	label := limit.ModelLabel
	if label == "" {
		label = limit.Model
	}
	if label == "" {
		label = "Model"
	}
	if len(label) <= orgSpendLimitModelLabelMax {
		return label
	}
	return label[:orgSpendLimitModelLabelMax-1] + "…"
}

func orgSpendLimitPercent(limit OrgLimit) float64 {
	if limit.Limit <= 0 {
		return 100
	}
	return math.Min(100, limit.Spent/limit.Limit*100)
}

func orgSpendLimitReached(limit OrgLimit) bool {
	return limit.Exceeded || limit.Limit <= 0
}

var orgLimitResetSuffix = map[string]string{
	"daily":   "/day",
	"weekly":  "/week",
	"monthly": "/mo",
	"total":   "",
}

func orgSpendLimitDetail(limit OrgLimit, now time.Time) string {
	detail := FormatCredits(limit.Spent) + " of " + FormatCredits(limit.Limit) + orgLimitResetSuffix[limit.ResetInterval]
	if orgSpendLimitReached(limit) {
		detail += " · limit reached"
	}
	if resetAt, ok := parseTime(limit.ResetAt); ok && resetAt.After(now) {
		detail += " · resets in " + FormatDuration(resetAt.Sub(now))
	}
	return detail
}

// FormatDuration renders a duration the way the CLI does: the largest two
// units, rounded up to a whole minute and never shorter than "1m".
func FormatDuration(d time.Duration) string {
	totalMinutes := int(math.Max(1, math.Ceil(d.Minutes())))
	days := totalMinutes / 1440
	hours := totalMinutes % 1440 / 60
	minutes := totalMinutes % 60
	switch {
	case days > 0 && hours > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case days > 0:
		return fmt.Sprintf("%dd", days)
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

// --- individual endpoint renderers ---

// FormatWhoami renders account identity as aligned key/value lines.
func FormatWhoami(whoami *Whoami) string {
	if whoami == nil {
		return ""
	}
	var b strings.Builder
	field(&b, "userName", whoami.User.UserName)
	field(&b, "name", whoami.User.Name)
	field(&b, "email", whoami.User.Email)
	field(&b, "userId", whoami.User.ID)
	if whoami.Org != nil {
		field(&b, "org", whoami.Org.Login+" ("+whoami.Org.ID+")")
	} else {
		field(&b, "org", "personal")
	}
	if len(whoami.OrgLimits) > 0 {
		field(&b, "limits", strconv.Itoa(len(whoami.OrgLimits)))
		for _, limit := range whoami.OrgLimits {
			field(&b, "  "+orgSpendLimitLabel(limit), orgSpendLimitDetail(limit, time.Now()))
		}
	}
	return b.String()
}

// FormatCreditsView renders the credit balance and rate-limit windows.
func FormatCreditsView(credits *Credits) string {
	if credits == nil {
		return ""
	}
	var b strings.Builder
	balance := credits.Credits
	field(&b, "monthly", FormatCredits(balance.MonthlyCredits))
	field(&b, "purchased", FormatCredits(balance.PurchasedCredits))
	field(&b, "free", FormatCredits(balance.FreeCredits))
	field(&b, "remaining", FormatCredits(balance.MonthlyCredits+balance.PurchasedCredits+balance.FreeCredits))
	threshold := FormatCredits(balance.CreditThreshold)
	if balance.BelowThreshold {
		threshold += " (below threshold)"
	}
	field(&b, "threshold", threshold)
	if windows := credits.WindowLimits; windows != nil {
		if windows.FiveHour != nil {
			field(&b, "5-hour", formatWindowSpan(windows.FiveHour, time.Now()))
		}
		if windows.Weekly != nil {
			field(&b, "weekly", formatWindowSpan(windows.Weekly, time.Now()))
		}
		field(&b, "limited", strconv.FormatBool(windows.Limited))
	}
	return b.String()
}

// FormatSubscriptionView renders the current subscription.
func FormatSubscriptionView(sub *Subscription) string {
	if sub == nil {
		return "no subscription\n"
	}
	var b strings.Builder
	field(&b, "id", sub.ID)
	field(&b, "status", sub.Status)
	plan := sub.PlanID
	if info := GetPlanInfo(sub.PlanID); info != nil {
		plan = fmt.Sprintf("%s (%s, %s credits/mo)", info.ID, info.Name, formatFloat(info.MonthlyCredits))
	}
	field(&b, "plan", plan)
	field(&b, "quantity", strconv.Itoa(sub.Quantity))
	field(&b, "cancelAtPeriodEnd", strconv.FormatBool(sub.CancelAtPeriodEnd))
	if sub.OrgID != nil {
		field(&b, "orgId", *sub.OrgID)
	} else {
		field(&b, "orgId", "personal")
	}
	if sub.CurrentPeriodStart != "" || sub.CurrentPeriodEnd != "" {
		field(&b, "currentPeriod", sub.CurrentPeriodStart+" → "+sub.CurrentPeriodEnd)
	}
	if end := sub.CurrentPeriodEnd; end != "" {
		if days := daysRemaining(end, time.Now()); days != nil {
			field(&b, "daysLeft", strconv.Itoa(*days))
		}
	}
	return b.String()
}

// FormatSummaryView renders the usage summary totals.
func FormatSummaryView(summary *UsageSummary) string {
	if summary == nil {
		return ""
	}
	var b strings.Builder
	field(&b, "requests", groupDigits(int64(summary.TotalCount)))
	field(&b, "completed", groupDigits(int64(summary.CompletedCount)))
	field(&b, "failed", groupDigits(int64(summary.FailedCount)))
	field(&b, "successRate", trimFloat(summary.SuccessRate)+"%")
	field(&b, "totalCost", FormatCredits(summary.TotalCost))
	field(&b, "averageCost", FormatCost(summary.AverageCost))
	field(&b, "tokensIn", groupDigits(summary.TotalTokensIn))
	field(&b, "tokensOut", groupDigits(summary.TotalTokensOut))
	field(&b, "tokens", groupDigits(summary.TotalTokens))
	field(&b, "credits", FormatCredits(summary.TotalCredits))
	field(&b, "monthlyCredits", FormatCredits(summary.TotalMonthlyCredits))
	field(&b, "purchasedCredits", FormatCredits(summary.TotalPurchasedCredits))
	field(&b, "freeCredits", FormatCredits(summary.TotalFreeCredits))
	field(&b, "periodBasis", summary.PeriodBasis)
	return b.String()
}

// FormatNamespacesView renders the account's namespaces.
func FormatNamespacesView(namespaces *Namespaces) string {
	if namespaces == nil {
		return ""
	}
	var b strings.Builder
	field(&b, "type", namespaces.Type)
	field(&b, "userName", namespaces.User.UserName)
	field(&b, "orgs", strconv.Itoa(len(namespaces.Orgs)))
	for _, org := range namespaces.Orgs {
		field(&b, "  "+org.Login, org.ID)
	}
	return b.String()
}

// --- shared helpers ---

type painter struct{ color bool }

func (p painter) paint(code, s string) string {
	if !p.color || code == "" || s == "" {
		return s
	}
	return code + s + ansiReset
}

func field(b *strings.Builder, key, value string) {
	fmt.Fprintf(b, "%-18s %s\n", key, value)
}

func formatWindowSpan(window *WindowSpan, now time.Time) string {
	percent := 0.0
	if window.Cap > 0 {
		percent = math.Min(100, window.Used/window.Cap*100)
	}
	out := fmt.Sprintf("%s / %s (%.0f%%)", formatFloat(window.Used), formatFloat(window.Cap), percent)
	if resetAt := windowResetTime(window.ResetAt); resetAt.After(now) {
		out += " · resets in " + FormatDuration(resetAt.Sub(now))
	}
	if window.Exceeded {
		out += " · exceeded"
	}
	return out
}

func subscriptionStatus(sub *Subscription) string {
	if sub == nil {
		return ""
	}
	return sub.Status
}

func usagePercentOf(used, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return math.Min(used/total*100, 100)
}

func usageColorCode(percent float64) string {
	switch {
	case percent >= 80:
		return ansiRed
	case percent >= 50:
		return ansiYellow
	default:
		return ansiGreen
	}
}

func daysColorCode(days int) string {
	switch renewalUrgency(days) {
	case "critical":
		return ansiRed
	case "warning":
		return ansiYellow
	default:
		return ""
	}
}

func renewalUrgency(days int) string {
	switch {
	case days < 3:
		return "critical"
	case days < 7:
		return "warning"
	default:
		return "normal"
	}
}

func daysRemaining(periodEnd string, now time.Time) *int {
	end, ok := parseTime(periodEnd)
	if !ok {
		return nil
	}
	days := int(math.Ceil(end.Sub(now).Hours() / 24))
	if days < 0 {
		days = 0
	}
	return &days
}

func parseTime(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, true
	}
	if ms, err := strconv.ParseFloat(value, 64); err == nil {
		return time.UnixMilli(int64(ms)), true
	}
	return time.Time{}, false
}

// windowResetTime converts a unix-millisecond reset stamp to a time. A zero or
// missing stamp yields the zero time, which never counts as a pending reset.
func windowResetTime(resetAt float64) time.Time {
	if resetAt <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(int64(resetAt))
}

func studioUsageURL(whoami *Whoami, host string) string {
	name := ""
	if whoami.Org != nil {
		name = whoami.Org.Login
	}
	if name == "" {
		name = whoami.User.UserName
	}
	if name == "" {
		return ""
	}
	return strings.TrimRight(host, "/") + "/" + name + "/settings/usage"
}

func stripScheme(url string) string {
	url = strings.TrimPrefix(url, "https://")
	return strings.TrimPrefix(url, "http://")
}

func groupDigits(n int64) string {
	negative := n < 0
	digits := strconv.FormatInt(n, 10)
	if negative {
		digits = digits[1:]
	}
	var out strings.Builder
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteByte(digits[i])
	}
	if negative {
		return "-" + out.String()
	}
	return out.String()
}

// formatFloat renders a credit amount rounded to cents, without a currency
// symbol or trailing zeros, for the column-oriented endpoint views.
func formatFloat(value float64) string {
	return strconv.FormatFloat(math.Round(value*100)/100, 'f', -1, 64)
}

// FormatCost renders a dollar amount that stays readable below one cent, where
// the CLI's two-decimal FormatCredits would collapse to $0.00.
func FormatCost(value float64) string {
	return "$" + strconv.FormatFloat(math.Round(value*10000)/10000, 'f', -1, 64)
}

// trimFloat renders a float with the shortest representation that round-trips.
func trimFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
