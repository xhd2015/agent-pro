package commandcode

import (
	"encoding/json"
	"sort"
	"strings"
)

// API paths. These are the read-only alpha endpoints the cmd CLI's usage
// overlay calls.
const (
	PathWhoami        = "/alpha/whoami"
	PathCredits       = "/alpha/billing/credits"
	PathSubscriptions = "/alpha/billing/subscriptions"
	PathUsageSummary  = "/alpha/usage/summary"
	PathNamespaces    = "/alpha/namespaces"
)

// User is the account identity embedded in several responses.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	UserName string `json:"userName"`
}

// Org identifies a team namespace.
type Org struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

// OrgLimit is one spend-limit row carried by whoami's orgLimits.
type OrgLimit struct {
	Scope         string  `json:"scope"`
	Model         string  `json:"model"`
	ModelLabel    string  `json:"modelLabel"`
	Limit         float64 `json:"limit"`
	Spent         float64 `json:"spent"`
	Exceeded      bool    `json:"exceeded"`
	ResetInterval string  `json:"resetInterval"`
	ResetAt       string  `json:"resetAt"`
}

// Whoami is the response of GET /alpha/whoami?limits=1.
type Whoami struct {
	Success   bool       `json:"success"`
	User      User       `json:"user"`
	Org       *Org       `json:"org"`
	OrgLimits []OrgLimit `json:"orgLimits,omitempty"`
}

// CreditBalance is the credits object of GET /alpha/billing/credits.
type CreditBalance struct {
	BelowThreshold   bool    `json:"belowThreshold"`
	CreditThreshold  float64 `json:"creditThreshold"`
	MonthlyCredits   float64 `json:"monthlyCredits"`
	PurchasedCredits float64 `json:"purchasedCredits"`
	FreeCredits      float64 `json:"freeCredits"`
}

// WindowSpan is one rate-limit window (fiveHour or weekly). ResetAt is a
// unix-millisecond timestamp, or 0 when the window has no pending reset.
type WindowSpan struct {
	Used     float64 `json:"used"`
	Cap      float64 `json:"cap"`
	Exceeded bool    `json:"exceeded"`
	ResetAt  float64 `json:"resetAt"`
}

// WindowLimits is the windowLimits object of GET /alpha/billing/credits.
type WindowLimits struct {
	Limited  bool        `json:"limited"`
	Exceeded *bool       `json:"exceeded"`
	FiveHour *WindowSpan `json:"fiveHour"`
	Weekly   *WindowSpan `json:"weekly"`
}

// Credits is the response of GET /alpha/billing/credits.
type Credits struct {
	Credits      CreditBalance `json:"credits"`
	WindowLimits *WindowLimits `json:"windowLimits"`
}

// Subscription is one subscription record.
type Subscription struct {
	ID                 string          `json:"id"`
	Status             string          `json:"status"`
	UserID             string          `json:"userId"`
	OrgID              *string         `json:"orgId"`
	CreatedAt          string          `json:"createdAt"`
	PriceID            string          `json:"priceId"`
	Metadata           json.RawMessage `json:"metadata"`
	Quantity           int             `json:"quantity"`
	CancelAtPeriodEnd  bool            `json:"cancelAtPeriodEnd"`
	CurrentPeriodStart string          `json:"currentPeriodStart"`
	CurrentPeriodEnd   string          `json:"currentPeriodEnd"`
	EndedAt            *string         `json:"endedAt"`
	CancelAt           *string         `json:"cancelAt"`
	CanceledAt         *string         `json:"canceledAt"`
	PlanID             string          `json:"planId"`
	PendingPhase       json.RawMessage `json:"pendingPhase"`
}

// SubscriptionResponse is the envelope of GET /alpha/billing/subscriptions.
type SubscriptionResponse struct {
	Success bool          `json:"success"`
	Data    *Subscription `json:"data"`
}

// UsageSummary is the response of GET /alpha/usage/summary. The numbers cover
// the billing period when since is the subscription's currentPeriodStart.
type UsageSummary struct {
	TotalCount            int     `json:"totalCount"`
	TotalCost             float64 `json:"totalCost"`
	AverageCost           float64 `json:"averageCost"`
	SuccessRate           float64 `json:"successRate"`
	CompletedCount        int     `json:"completedCount"`
	FailedCount           int     `json:"failedCount"`
	TotalTokensIn         int64   `json:"totalTokensIn"`
	TotalTokensOut        int64   `json:"totalTokensOut"`
	TotalTokens           int64   `json:"totalTokens"`
	TotalCredits          float64 `json:"totalCredits"`
	TotalFreeCredits      float64 `json:"totalFreeCredits"`
	TotalMonthlyCredits   float64 `json:"totalMonthlyCredits"`
	TotalPurchasedCredits float64 `json:"totalPurchasedCredits"`
	PeriodBasis           string  `json:"periodBasis"`
}

// Namespaces is the response of GET /alpha/namespaces.
type Namespaces struct {
	Success bool   `json:"success"`
	Type    string `json:"type"`
	User    User   `json:"user"`
	Org     *Org   `json:"org,omitempty"`
	Orgs    []Org  `json:"orgs"`
}

// PlanCredits maps a plan id to its monthly credit allowance.
var PlanCredits = map[string]float64{
	"individual-go":       10,
	"individual-goat":     70,
	"individual-pro":      30,
	"individual-pro-v1":   80,
	"individual-provider": 15,
	"individual-max":      150,
	"individual-ultra":    300,
	"teams-pro":           40,
}

// PlanNames maps a plan id to its display name.
var PlanNames = map[string]string{
	"individual-go":       "Go",
	"individual-goat":     "GOAT",
	"individual-pro":      "Pro",
	"individual-pro-v1":   "Pro",
	"individual-provider": "Provider",
	"individual-max":      "Max",
	"individual-ultra":    "Ultra",
	"teams-pro":           "Teams Pro",
}

// planPrefixes orders plan ids longest first so a prefix match prefers the
// most specific plan (individual-pro-v1 before individual-pro).
var planPrefixes = func() []string {
	prefixes := make([]string, 0, len(PlanCredits))
	for id := range PlanCredits {
		prefixes = append(prefixes, id)
	}
	sort.Slice(prefixes, func(i, j int) bool {
		if len(prefixes[i]) != len(prefixes[j]) {
			return len(prefixes[i]) > len(prefixes[j])
		}
		return prefixes[i] < prefixes[j]
	})
	return prefixes
}()

// Plan is a resolved plan identity.
type Plan struct {
	ID             string
	Name           string
	MonthlyCredits float64
}

// GetPlanInfo resolves a planId to its display name and credit allowance.
// Matching lowercases the id, maps underscores to dashes, and takes the
// longest known prefix. It returns nil for empty or unknown ids.
func GetPlanInfo(planID string) *Plan {
	if strings.TrimSpace(planID) == "" {
		return nil
	}
	normalized := strings.ReplaceAll(strings.ToLower(planID), "_", "-")
	for _, prefix := range planPrefixes {
		if strings.HasPrefix(normalized, prefix) {
			name := PlanNames[prefix]
			if name == "" {
				name = prefix
			}
			return &Plan{ID: prefix, Name: name, MonthlyCredits: PlanCredits[prefix]}
		}
	}
	return nil
}
