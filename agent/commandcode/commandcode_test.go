package commandcode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- auth ---

func TestReadAuthPrefersEnvKey(t *testing.T) {
	home := t.TempDir()
	writeAuthFile(t, home, `{"apiKey":"from-file","userName":"file-user"}`)
	t.Setenv(APIKeyEnvVar, "from-env")

	auth, err := ReadAuth(home)
	if err != nil {
		t.Fatalf("ReadAuth: %v", err)
	}
	if auth.APIKey != "from-env" {
		t.Fatalf("APIKey=%q want from-env", auth.APIKey)
	}
	if auth.Source != "env "+APIKeyEnvVar {
		t.Fatalf("Source=%q", auth.Source)
	}
}

func TestReadAuthReadsFileKey(t *testing.T) {
	home := t.TempDir()
	writeAuthFile(t, home, `{"apiKey":"from-file","userName":"file-user"}`)
	t.Setenv(APIKeyEnvVar, "")

	auth, err := ReadAuth(home)
	if err != nil {
		t.Fatalf("ReadAuth: %v", err)
	}
	if auth.APIKey != "from-file" {
		t.Fatalf("APIKey=%q want from-file", auth.APIKey)
	}
	if auth.Source != AuthPath(home) {
		t.Fatalf("Source=%q want %q", auth.Source, AuthPath(home))
	}
}

func TestReadAuthMissingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv(APIKeyEnvVar, "")

	_, err := ReadAuth(home)
	if err == nil {
		t.Fatal("ReadAuth succeeded without credentials")
	}
	if !strings.Contains(err.Error(), "not authenticated") || !strings.Contains(err.Error(), APIKeyEnvVar) {
		t.Fatalf("err=%v", err)
	}
}

func TestReadAuthEmptyKey(t *testing.T) {
	home := t.TempDir()
	writeAuthFile(t, home, `{"apiKey":"   ","userName":"u"}`)
	t.Setenv(APIKeyEnvVar, "")

	_, err := ReadAuth(home)
	if err == nil || !strings.Contains(err.Error(), "has no apiKey") {
		t.Fatalf("err=%v", err)
	}
}

// --- request shape ---

func TestEndpointOmitsEmptyParams(t *testing.T) {
	client := &Client{BaseURL: "https://api.example.test/"}
	cases := []struct {
		name   string
		path   string
		params map[string]string
		want   string
	}{
		{"nil params", PathNamespaces, nil, "https://api.example.test" + PathNamespaces},
		{"empty values dropped", PathCredits, map[string]string{"orgId": ""}, "https://api.example.test" + PathCredits},
		{"blank value dropped", PathCredits, map[string]string{"orgId": "  "}, "https://api.example.test" + PathCredits},
		{"value kept", PathCredits, map[string]string{"orgId": "org_1"}, "https://api.example.test" + PathCredits + "?orgId=org_1"},
		{"limits kept, org dropped", PathWhoami, map[string]string{"limits": "1", "orgId": ""}, "https://api.example.test" + PathWhoami + "?limits=1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := client.Endpoint(tc.path, tc.params); got != tc.want {
				t.Fatalf("Endpoint=%q want %q", got, tc.want)
			}
		})
	}
}

func TestResolveBaseURL(t *testing.T) {
	t.Setenv(sandboxEnvVar, "")
	t.Setenv(apiURLEnvVar, "")
	if got := ResolveBaseURL(""); got != DefaultBaseURL {
		t.Fatalf("default=%q", got)
	}
	if got := ResolveBaseURL("http://127.0.0.1:9/"); got != "http://127.0.0.1:9" {
		t.Fatalf("explicit=%q", got)
	}
	t.Setenv(sandboxEnvVar, "true")
	t.Setenv(apiURLEnvVar, "http://sandbox.test")
	if got := ResolveBaseURL(""); got != "http://sandbox.test" {
		t.Fatalf("sandbox=%q", got)
	}
	t.Setenv(sandboxEnvVar, "false")
	if got := ResolveBaseURL(""); got != DefaultBaseURL {
		t.Fatalf("non-sandbox=%q", got)
	}
}

// --- request headers and error mapping ---

func TestGetSendsBearerTokenAndDecodes(t *testing.T) {
	var gotAuth, gotAgent, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAgent = r.Header.Get("User-Agent")
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"credits":{"monthlyCredits":1.5}}`))
	}))
	defer server.Close()

	client := testClient(t, server.URL, "test-key")
	credits, err := client.Credits(context.Background(), "org_9")
	if err != nil {
		t.Fatalf("Credits: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("Authorization=%q", gotAuth)
	}
	if gotAgent != UserAgent {
		t.Fatalf("User-Agent=%q want %q", gotAgent, UserAgent)
	}
	if gotQuery != "orgId=org_9" {
		t.Fatalf("query=%q", gotQuery)
	}
	if credits.Credits.MonthlyCredits != 1.5 {
		t.Fatalf("credits=%+v", credits.Credits)
	}
}

func TestGetUnauthorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"UNAUTHORIZED","status":401,"message":"invalid key"}}`))
	}))
	defer server.Close()

	client := testClient(t, server.URL, "bad-key")
	_, err := client.Whoami(context.Background(), "")
	if err == nil {
		t.Fatal("Whoami succeeded with a rejected key")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type %T: %v", err, err)
	}
	if !apiErr.Unauthorized() || apiErr.Status != http.StatusUnauthorized || apiErr.Code != "UNAUTHORIZED" {
		t.Fatalf("apiErr=%+v", apiErr)
	}
	if !strings.Contains(apiErr.Error(), "invalid key") {
		t.Fatalf("apiErr.Error()=%q", apiErr.Error())
	}
}

func TestGetNonEnvelopeErrorFallsBackToBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream exploded"))
	}))
	defer server.Close()

	client := testClient(t, server.URL, "key")
	_, err := client.Whoami(context.Background(), "")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type %T: %v", err, err)
	}
	if apiErr.Unauthorized() {
		t.Fatalf("502 marked unauthorized: %+v", apiErr)
	}
	if !strings.Contains(apiErr.Error(), "upstream exploded") {
		t.Fatalf("apiErr.Error()=%q", apiErr.Error())
	}
}

func TestGetNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()

	client := testClient(t, url, "key")
	_, err := client.Whoami(context.Background(), "")
	if _, ok := err.(*NetworkError); !ok {
		t.Fatalf("err type %T: %v", err, err)
	}
}

// --- usage orchestration ---

func TestFetchUsageThreadsOrgAndPeriod(t *testing.T) {
	var summaryQuery string
	paths := make([]string, 0, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case PathWhoami:
			if r.URL.Query().Get("orgId") != "" {
				t.Errorf("whoami sent orgId=%q", r.URL.Query().Get("orgId"))
			}
			_, _ = w.Write([]byte(`{"success":true,"user":{"userName":"alice"},"org":{"id":"org_7","login":"acme"}}`))
		case PathCredits:
			if r.URL.Query().Get("orgId") != "org_7" {
				t.Errorf("credits orgId=%q want org_7", r.URL.Query().Get("orgId"))
			}
			_, _ = w.Write([]byte(`{"credits":{"monthlyCredits":3}}`))
		case PathSubscriptions:
			_, _ = w.Write([]byte(`{"success":true,"data":{"status":"active","planId":"individual-go","currentPeriodStart":"2026-08-14T00:05:12.000Z","currentPeriodEnd":"2026-09-14T00:05:12.000Z"}}`))
		case PathUsageSummary:
			summaryQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{"totalCount":7}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := testClient(t, server.URL, "key")
	data, err := client.FetchUsage(context.Background())
	if err != nil {
		t.Fatalf("FetchUsage: %v", err)
	}
	if len(data.Errors) != 0 {
		t.Fatalf("Errors=%v", data.Errors)
	}
	if data.Whoami == nil || data.Credits == nil || data.Subscription == nil || data.Summary == nil {
		t.Fatalf("incomplete payload: %+v", data)
	}
	if len(paths) != 4 {
		t.Fatalf("paths=%v", paths)
	}
	if !strings.Contains(summaryQuery, "orgId=org_7") || !strings.Contains(summaryQuery, "since=2026-08-14T00%3A05%3A12.000Z") {
		t.Fatalf("summary query=%q", summaryQuery)
	}
}

func TestFetchUsageOverridesWin(t *testing.T) {
	var summaryQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PathWhoami:
			_, _ = w.Write([]byte(`{"success":true,"user":{"userName":"alice"},"org":{"id":"org_7"}}`))
		case PathCredits, PathSubscriptions:
			if got := r.URL.Query().Get("orgId"); got != "org_pinned" {
				t.Errorf("%s orgId=%q want org_pinned", r.URL.Path, got)
			}
			if r.URL.Path == PathSubscriptions {
				_, _ = w.Write([]byte(`{"success":true,"data":{"status":"active","currentPeriodStart":"2026-08-14T00:05:12.000Z"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"credits":{"monthlyCredits":1}}`))
		case PathUsageSummary:
			summaryQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{"totalCount":1}`))
		}
	}))
	defer server.Close()

	client := testClient(t, server.URL, "key")
	_, err := client.FetchUsageWithOptions(context.Background(), UsageOptions{
		OrgID: "org_pinned",
		Since: "2026-01-02T03:04:05Z",
	})
	if err != nil {
		t.Fatalf("FetchUsageWithOptions: %v", err)
	}
	if !strings.Contains(summaryQuery, "orgId=org_pinned") || !strings.Contains(summaryQuery, "since=2026-01-02T03%3A04%3A05Z") {
		t.Fatalf("summary query=%q", summaryQuery)
	}
}

func TestFetchUsageCollectsPartialFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PathWhoami:
			_, _ = w.Write([]byte(`{"success":true,"user":{"userName":"alice"}}`))
		case PathCredits:
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"success":false,"error":{"code":"UNAUTHORIZED","status":401,"message":"expired"}}`))
		case PathSubscriptions:
			_, _ = w.Write([]byte(`{"success":true,"data":{"status":"active","planId":"individual-go"}}`))
		case PathUsageSummary:
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"success":false,"error":{"code":"UPSTREAM","status":502,"message":"boom"}}`))
		}
	}))
	defer server.Close()

	client := testClient(t, server.URL, "key")
	data, err := client.FetchUsage(context.Background())
	if err != nil {
		t.Fatalf("FetchUsage: %v", err)
	}
	if data.Credits != nil {
		t.Fatalf("Credits should be nil on failure: %+v", data.Credits)
	}
	if data.Subscription == nil || data.Summary != nil {
		t.Fatalf("payload=%+v", data)
	}
	if len(data.Errors) != 2 {
		t.Fatalf("Errors=%v", data.Errors)
	}
	if data.Errors[0] != ErrSessionExpired {
		t.Fatalf("credits error=%q want %q", data.Errors[0], ErrSessionExpired)
	}
	if !strings.Contains(data.Errors[1], PathUsageSummary) {
		t.Fatalf("summary error=%q", data.Errors[1])
	}
}

// Only a 401 is an expired session. A 403 means the credential is fine but the
// scope is not permitted, and the API's own explanation is what helps there.
func TestUsageErrorWording(t *testing.T) {
	cases := []struct {
		name          string
		status        int
		body          string
		wantSubstr    []string
		wantAbsent    []string
		namesEndpoint bool
	}{
		{
			name:       "unauthorized is an expired session",
			status:     http.StatusUnauthorized,
			body:       `{"success":false,"error":{"code":"UNAUTHORIZED","status":401,"message":"invalid api key"}}`,
			wantSubstr: []string{ErrSessionExpired},
		},
		{
			name:   "forbidden explains the missing permission",
			status: http.StatusForbidden,
			body:   `{"success":false,"error":{"code":"FORBIDDEN","status":403,"message":"You do not have permission to view usage for this organization"}}`,
			wantSubstr: []string{
				PathUsageSummary,
				"You do not have permission to view usage for this organization",
				"FORBIDDEN 403",
			},
			wantAbsent:    []string{ErrSessionExpired},
			namesEndpoint: true,
		},
		{
			name:          "bad request keeps the validation hint",
			status:        http.StatusBadRequest,
			body:          `{"success":false,"error":{"code":"BAD_REQUEST","status":400,"message":"Invalid UUID at \"orgId\""}}`,
			wantSubstr:    []string{PathUsageSummary, "Invalid UUID", "BAD_REQUEST 400"},
			namesEndpoint: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case PathWhoami:
					_, _ = w.Write([]byte(`{"success":true,"user":{"userName":"alice"}}`))
				case PathCredits:
					_, _ = w.Write([]byte(`{"credits":{"monthlyCredits":2.4}}`))
				case PathSubscriptions:
					_, _ = w.Write([]byte(`{"success":true,"data":{"status":"active","planId":"individual-go"}}`))
				default:
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(tc.body))
				}
			}))
			defer server.Close()

			data, err := testClient(t, server.URL, "key").FetchUsage(context.Background())
			if err != nil {
				t.Fatalf("FetchUsage: %v", err)
			}
			if len(data.Errors) != 1 {
				t.Fatalf("Errors=%v, want one summary failure", data.Errors)
			}
			msg := data.Errors[0]
			for _, want := range tc.wantSubstr {
				if !strings.Contains(msg, want) {
					t.Fatalf("error=%q missing %q", msg, want)
				}
			}
			for _, unwanted := range tc.wantAbsent {
				if strings.Contains(msg, unwanted) {
					t.Fatalf("error=%q should not contain %q", msg, unwanted)
				}
			}
			if tc.namesEndpoint {
				if got := strings.Count(msg, PathUsageSummary); got != 1 {
					t.Fatalf("error=%q names the endpoint %d times, want 1", msg, got)
				}
			}
		})
	}
}

func TestForbiddenIsNotAnAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"FORBIDDEN","status":403,"message":"You do not have permission to view usage for this organization"}}`))
	}))
	defer server.Close()

	_, err := testClient(t, server.URL, "key").Whoami(context.Background(), "")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type %T: %v", err, err)
	}
	if apiErr.Unauthorized() {
		t.Fatalf("403 marked unauthorized: %+v", apiErr)
	}
	if !strings.Contains(apiErr.Error(), "do not have permission") {
		t.Fatalf("apiErr.Error()=%q", apiErr.Error())
	}
}

func TestFetchUsageStopsWhenWhoamiFails(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"UNAUTHORIZED","status":401,"message":"nope"}}`))
	}))
	defer server.Close()

	client := testClient(t, server.URL, "key")
	data, err := client.FetchUsage(context.Background())
	if err != nil {
		t.Fatalf("FetchUsage: %v", err)
	}
	if data.Whoami != nil || len(data.Errors) != 1 || data.Errors[0] != ErrSessionExpired {
		t.Fatalf("payload=%+v", data)
	}
	if hits != 1 {
		t.Fatalf("server hits=%d want 1", hits)
	}
}

// --- plan resolution ---

func TestGetPlanInfo(t *testing.T) {
	cases := []struct {
		planID  string
		wantNil bool
		name    string
		credits float64
	}{
		{planID: "", wantNil: true},
		{planID: "unknown-plan", wantNil: true},
		{planID: "individual-go", name: "Go", credits: 10},
		{planID: "individual-goat", name: "GOAT", credits: 70},
		{planID: "individual-pro", name: "Pro", credits: 30},
		// individual-pro-v1 must win over individual-pro (longest prefix).
		{planID: "individual-pro-v1", name: "Pro", credits: 80},
		{planID: "individual_goat", name: "GOAT", credits: 70},
		{planID: "INDIVIDUAL-MAX", name: "Max", credits: 150},
		{planID: "teams-pro", name: "Teams Pro", credits: 40},
	}
	for _, tc := range cases {
		t.Run(tc.planID, func(t *testing.T) {
			plan := GetPlanInfo(tc.planID)
			if tc.wantNil {
				if plan != nil {
					t.Fatalf("plan=%+v want nil", plan)
				}
				return
			}
			if plan == nil {
				t.Fatal("plan=nil")
			}
			if plan.Name != tc.name || plan.MonthlyCredits != tc.credits {
				t.Fatalf("plan=%+v want name=%s credits=%v", plan, tc.name, tc.credits)
			}
		})
	}
}

// --- view math ---

func TestProjectUsageViewActivePlan(t *testing.T) {
	now := time.Date(2026, 9, 9, 0, 5, 12, 0, time.UTC)
	data := &UsageData{
		Whoami: &Whoami{User: User{UserName: "alice"}, Success: true},
		Credits: &Credits{
			Credits: CreditBalance{MonthlyCredits: 2.408591571},
			WindowLimits: &WindowLimits{Limited: true, Weekly: &WindowSpan{
				Used: 0.480063398, Cap: 6, ResetAt: float64(now.Add(7*time.Hour + 4*time.Minute).UnixMilli()),
			}},
		},
		Subscription: &Subscription{
			Status:             "active",
			PlanID:             "individual-go",
			CurrentPeriodStart: "2026-08-14T00:05:12.000Z",
			CurrentPeriodEnd:   "2026-09-14T00:05:12.000Z",
		},
		Summary: &UsageSummary{TotalCount: 1040, TotalCost: 7.5914084289999995},
		Errors:  []string{},
	}

	view := ProjectUsageView(data, now)
	if !view.HasBillingData {
		t.Fatal("HasBillingData=false")
	}
	if view.Plan == nil || view.Plan.Name != "Go" {
		t.Fatalf("Plan=%+v", view.Plan)
	}
	// Pool is the plan allowance (10), not the remaining monthly credits.
	if view.Credits.TotalPool != 10 {
		t.Fatalf("TotalPool=%v want 10", view.Credits.TotalPool)
	}
	if view.Credits.TotalRemaining != 2.408591571 {
		t.Fatalf("TotalRemaining=%v", view.Credits.TotalRemaining)
	}
	if got := int(view.Credits.UsagePercent + 0.5); got != 76 {
		t.Fatalf("UsagePercent=%v want ~75.91", view.Credits.UsagePercent)
	}
	if view.DaysLeft == nil || *view.DaysLeft != 5 {
		t.Fatalf("DaysLeft=%v want 5", view.DaysLeft)
	}
	if view.UsageURL != DefaultStudioHost+"/alice/settings/usage" {
		t.Fatalf("UsageURL=%q", view.UsageURL)
	}
	if view.UsageURLDisplay != "commandcode.ai/alice/settings/usage" {
		t.Fatalf("UsageURLDisplay=%q", view.UsageURLDisplay)
	}
}

func TestProjectUsageViewNonActiveUsesSpendAndRemaining(t *testing.T) {
	now := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	data := &UsageData{
		Whoami: &Whoami{User: User{UserName: "alice"}},
		Credits: &Credits{Credits: CreditBalance{
			MonthlyCredits: 4, PurchasedCredits: 1, FreeCredits: 1,
		}},
		Subscription: &Subscription{Status: "canceled", PlanID: "individual-go"},
		Summary:      &UsageSummary{TotalCost: 4},
	}
	view := ProjectUsageView(data, now)
	if view.Credits.TotalPool != 10 {
		t.Fatalf("TotalPool=%v want 10 (4 spent + 6 remaining)", view.Credits.TotalPool)
	}
	if view.Credits.UsagePercent != 40 {
		t.Fatalf("UsagePercent=%v want 40", view.Credits.UsagePercent)
	}
	if view.DaysLeft != nil {
		t.Fatalf("DaysLeft=%v want nil without a period end", *view.DaysLeft)
	}
}

func TestProjectUsageViewEmpty(t *testing.T) {
	view := ProjectUsageView(&UsageData{}, time.Now())
	if view.HasBillingData {
		t.Fatal("HasBillingData=true for an empty payload")
	}
	if view.Credits.HasCreditsInfo {
		t.Fatal("HasCreditsInfo=true for an empty payload")
	}
	if view.Credits.UsagePercent != 0 {
		t.Fatalf("UsagePercent=%v", view.Credits.UsagePercent)
	}
	if view.UsageURL != "" {
		t.Fatalf("UsageURL=%q", view.UsageURL)
	}
}

func TestProjectUsageViewOrgLoginWinsForURL(t *testing.T) {
	data := &UsageData{
		Whoami: &Whoami{
			User: User{UserName: "alice"},
			Org:  &Org{ID: "org_1", Login: "acme"},
		},
	}
	view := ProjectUsageView(data, time.Now())
	if view.UsageURL != DefaultStudioHost+"/acme/settings/usage" {
		t.Fatalf("UsageURL=%q", view.UsageURL)
	}
}

// --- bar math ---

func TestBuildBlockBar(t *testing.T) {
	cases := []struct {
		percent float64
		width   int
		filled  int
	}{
		{percent: 0, width: 30, filled: 0},
		{percent: 100, width: 30, filled: 30},
		{percent: 99.4, width: 30, filled: 29}, // just below full keeps one empty block
		{percent: 50, width: 30, filled: 15},
		{percent: 8.001, width: 30, filled: 2},
		{percent: 0.4, width: 30, filled: 0},
		{percent: 0.6, width: 30, filled: 1}, // nonzero above the guard shows one block
		{percent: 75.914, width: 30, filled: 23},
		{percent: 200, width: 10, filled: 10}, // clamped to 100
		{percent: -5, width: 10, filled: 0},   // clamped to 0
	}
	for _, tc := range cases {
		bar := BuildBlockBar(tc.percent, tc.width)
		if len([]rune(bar.Filled)) != tc.filled {
			t.Fatalf("percent=%v width=%d filled=%d want %d", tc.percent, tc.width, len([]rune(bar.Filled)), tc.filled)
		}
		if len([]rune(bar.Filled))+len([]rune(bar.Empty)) != tc.width {
			t.Fatalf("percent=%v bar width=%d want %d", tc.percent, len([]rune(bar.Filled))+len([]rune(bar.Empty)), tc.width)
		}
	}
}

func TestProgressBarWidth(t *testing.T) {
	cases := map[int]int{40: 30, 49: 39, 50: 35, 69: 54, 70: 30, 80: 30, 200: 30}
	for terminal, want := range cases {
		if got := ProgressBarWidth(terminal); got != want {
			t.Fatalf("ProgressBarWidth(%d)=%d want %d", terminal, got, want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{d: 7*time.Hour + 4*time.Minute, want: "7h 4m"},
		{d: 4*time.Hour + 2*time.Minute, want: "4h 2m"},
		{d: 6*24*time.Hour + 11*time.Hour, want: "6d 11h"},
		{d: 30 * 24 * time.Hour, want: "30d"},
		{d: 5 * time.Hour, want: "5h"},
		{d: time.Second, want: "1m"}, // floor of one minute
		{d: -time.Hour, want: "1m"},
	}
	for _, tc := range cases {
		if got := FormatDuration(tc.d); got != tc.want {
			t.Fatalf("FormatDuration(%v)=%q want %q", tc.d, got, tc.want)
		}
	}
}

func TestFormatCostKeepsSubCentValues(t *testing.T) {
	if got := FormatCost(0.0072994311817307695); got != "$0.0073" {
		t.Fatalf("FormatCost=%q", got)
	}
	if got := FormatCredits(7.5914084289999995); got != "$7.59" {
		t.Fatalf("FormatCredits=%q", got)
	}
}

// --- rendered overlay ---

func TestFormatUsageMatchesCLIOverlay(t *testing.T) {
	now := time.Date(2026, 9, 9, 0, 5, 12, 0, time.UTC)
	data := usageFixture(now)
	view := ProjectUsageView(data, now)

	got := FormatUsage(view, FormatOptions{Width: 80})
	want := strings.Join([]string{
		" USAGE  Go Plan · active",
		"",
		strings.Repeat("█", 23) + strings.Repeat("░", 7) + " 76% used",
		"Cycle: $2.41 left · 1,040 requests · 5 days to renewal",
		"",
		"Usage limits",
		"5-hour  " + strings.Repeat("░", 30) + " 0%",
		"",
		"Weekly  " + strings.Repeat("█", 2) + strings.Repeat("░", 28) + " 8% · resets in 7h 4m",
		"",
		"Full breakdown at commandcode.ai/alice/settings/usage",
		"",
	}, "\n")
	if got != want {
		t.Fatalf("usage overlay mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatUsageColorsAndWidth(t *testing.T) {
	now := time.Date(2026, 9, 9, 0, 5, 12, 0, time.UTC)
	view := ProjectUsageView(usageFixture(now), now)

	colored := FormatUsage(view, FormatOptions{Width: 80, Color: true})
	if !strings.Contains(colored, ansiBadge+" USAGE "+ansiReset+" ") {
		t.Fatalf("missing USAGE badge:\n%q", colored)
	}
	if !strings.Contains(colored, ansiGreen+"active"+ansiReset) {
		t.Fatalf("status not green:\n%q", colored)
	}
	// 75.9% crosses the 50% yellow threshold but not 80% red.
	if !strings.Contains(colored, ansiYellow) || strings.Contains(colored, ansiRed) {
		t.Fatalf("unexpected usage color:\n%q", colored)
	}

	// Bar width follows the CLI's terminal-width tiers; a zero width falls back
	// to the package default (80 columns -> 30).
	leadingBarWidth := func(width int) int {
		line := strings.SplitN(FormatUsage(view, FormatOptions{Width: width}), "\n", 4)[2]
		return len([]rune(strings.SplitN(line, " ", 2)[0]))
	}
	for _, tc := range []struct{ width, bar int }{
		{width: 80, bar: 30},
		{width: 60, bar: 45},
		{width: 40, bar: 30},
		{width: 0, bar: 30},
	} {
		if got := leadingBarWidth(tc.width); got != tc.bar {
			t.Fatalf("bar width at %d columns=%d want %d", tc.width, got, tc.bar)
		}
	}
}

func TestFormatUsageNoBillingData(t *testing.T) {
	data := &UsageData{Whoami: &Whoami{User: User{UserName: "alice"}}}
	got := FormatUsage(ProjectUsageView(data, time.Now()), FormatOptions{Width: 80})
	if !strings.Contains(got, " USAGE \n") {
		t.Fatalf("missing header:\n%s", got)
	}
	if !strings.Contains(got, "No billing data found.") {
		t.Fatalf("missing empty state:\n%s", got)
	}
	if !strings.Contains(got, "Visit Studio for usage details.") {
		t.Fatalf("missing Studio hint:\n%s", got)
	}
	if strings.Contains(got, "used") {
		t.Fatalf("empty state rendered a usage bar:\n%s", got)
	}
}

func TestFormatUsagePlanDetailsUnavailable(t *testing.T) {
	// Billing data exists but there is nothing spent or left to measure.
	data := &UsageData{
		Whoami:       &Whoami{User: User{UserName: "alice"}},
		Credits:      &Credits{Credits: CreditBalance{}},
		Subscription: &Subscription{Status: "active", PlanID: "individual-go"},
	}
	got := FormatUsage(ProjectUsageView(data, time.Now()), FormatOptions{Width: 80})
	if !strings.Contains(got, "Plan details unavailable") {
		t.Fatalf("missing fallback line:\n%s", got)
	}
}

func TestFormatUsageJSONKeys(t *testing.T) {
	now := time.Date(2026, 9, 9, 0, 5, 12, 0, time.UTC)
	raw, err := FormatUsageJSON(usageFixture(now))
	if err != nil {
		t.Fatalf("FormatUsageJSON: %v", err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"whoami", "credits", "subscription", "summary", "errors"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing key %q in:\n%s", key, raw)
		}
	}
	if len(decoded) != 5 {
		t.Fatalf("keys=%v", decoded)
	}
}

func TestFormatUsageJSONEmptyErrorsNotNil(t *testing.T) {
	raw, err := FormatUsageJSON(&UsageData{Errors: []string{}})
	if err != nil {
		t.Fatalf("FormatUsageJSON: %v", err)
	}
	if !strings.Contains(string(raw), `"errors": []`) {
		t.Fatalf("errors should render as []:\n%s", raw)
	}
}

// --- endpoint views ---

func TestFormatWhoami(t *testing.T) {
	got := FormatWhoami(&Whoami{
		User: User{ID: "u1", Name: "Alice", Email: "alice@example.test", UserName: "alice"},
	})
	for _, want := range []string{
		"userName           alice\n",
		"name               Alice\n",
		"email              alice@example.test\n",
		"userId             u1\n",
		"org                personal\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "limits") {
		t.Fatalf("limits row printed without limits:\n%s", got)
	}
}

func TestFormatWhoamiOrgLimits(t *testing.T) {
	got := FormatWhoami(&Whoami{
		User: User{UserName: "alice"},
		Org:  &Org{ID: "org_1", Login: "acme"},
		OrgLimits: []OrgLimit{
			{Scope: "org", Limit: 100, Spent: 25, ResetInterval: "monthly"},
			{Scope: "model", ModelLabel: "GPT-5", Limit: 10, Spent: 10, Exceeded: true, ResetInterval: "daily"},
		},
	})
	if !strings.Contains(got, "org                acme (org_1)\n") {
		t.Fatalf("missing org line:\n%s", got)
	}
	if !strings.Contains(got, "limits             2\n") {
		t.Fatalf("missing limits count:\n%s", got)
	}
	if !strings.Contains(got, "$25.00 of $100.00/mo") {
		t.Fatalf("missing org-wide row:\n%s", got)
	}
	if !strings.Contains(got, "$10.00 of $10.00/day · limit reached") {
		t.Fatalf("missing model row:\n%s", got)
	}
}

func TestFormatWhoamiTruncatesLongModelLabel(t *testing.T) {
	got := FormatWhoami(&Whoami{
		User: User{UserName: "alice"},
		OrgLimits: []OrgLimit{{
			Scope: "model", ModelLabel: "a-very-long-model-label-name", Limit: 1, Spent: 0,
		}},
	})
	if !strings.Contains(got, "a-very-long-model…") {
		t.Fatalf("label not truncated:\n%s", got)
	}
}

func TestFormatCreditsView(t *testing.T) {
	now := time.Now()
	got := FormatCreditsView(&Credits{
		Credits: CreditBalance{
			MonthlyCredits: 2.5, PurchasedCredits: 1, FreeCredits: 0.5,
			CreditThreshold: 1, BelowThreshold: true,
		},
		WindowLimits: &WindowLimits{
			Limited:  true,
			FiveHour: &WindowSpan{Used: 1, Cap: 3, ResetAt: float64(now.Add(4*time.Hour + 2*time.Minute).UnixMilli())},
			Weekly:   &WindowSpan{Used: 6, Cap: 6, Exceeded: true},
		},
	})
	for _, want := range []string{
		"monthly            $2.50\n",
		"purchased          $1.00\n",
		"free               $0.50\n",
		"remaining          $4.00\n",
		"threshold          $1.00 (below threshold)\n",
		"5-hour             1 / 3 (33%) · resets in 4h 2m\n",
		"weekly             6 / 6 (100%) · exceeded\n",
		"limited            true\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFormatSubscriptionView(t *testing.T) {
	got := FormatSubscriptionView(&Subscription{
		ID: "sub_1", Status: "active", PlanID: "individual-goat", Quantity: 1,
		CurrentPeriodStart: "2026-08-14T00:05:12.000Z",
		CurrentPeriodEnd:   "2026-09-14T00:05:12.000Z",
	})
	for _, want := range []string{
		"status             active\n",
		"plan               individual-goat (GOAT, 70 credits/mo)\n",
		"orgId              personal\n",
		"currentPeriod      2026-08-14T00:05:12.000Z → 2026-09-14T00:05:12.000Z\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "cancelAtPeriodEnd  true") {
		t.Fatalf("cancelAtPeriodEnd wrong:\n%s", got)
	}
}

func TestFormatSubscriptionViewNone(t *testing.T) {
	if got := FormatSubscriptionView(nil); got != "no subscription\n" {
		t.Fatalf("got=%q", got)
	}
}

func TestFormatSummaryView(t *testing.T) {
	got := FormatSummaryView(&UsageSummary{
		TotalCount: 1040, CompletedCount: 1040, SuccessRate: 100,
		TotalCost: 7.5914084289999995, AverageCost: 0.007299431181730769,
		TotalTokensIn: 22637783, TotalTokensOut: 221861, TotalTokens: 22859644,
		TotalCredits: 7.591408429, TotalMonthlyCredits: 7.591408429,
		PeriodBasis: "billing-period",
	})
	for _, want := range []string{
		"requests           1,040\n",
		"successRate        100%\n",
		"totalCost          $7.59\n",
		"averageCost        $0.0073\n",
		"tokensIn           22,637,783\n",
		"periodBasis        billing-period\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestFormatNamespacesView(t *testing.T) {
	got := FormatNamespacesView(&Namespaces{
		Type: "personal",
		User: User{UserName: "alice"},
		Orgs: []Org{{Login: "acme", ID: "org_1"}},
	})
	for _, want := range []string{
		"type               personal\n",
		"userName           alice\n",
		"orgs               1\n",
		"  acme             org_1\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestNilEndpointViewsRenderEmpty(t *testing.T) {
	if got := FormatWhoami(nil); got != "" {
		t.Fatalf("FormatWhoami(nil)=%q", got)
	}
	if got := FormatCreditsView(nil); got != "" {
		t.Fatalf("FormatCreditsView(nil)=%q", got)
	}
	if got := FormatSummaryView(nil); got != "" {
		t.Fatalf("FormatSummaryView(nil)=%q", got)
	}
	if got := FormatNamespacesView(nil); got != "" {
		t.Fatalf("FormatNamespacesView(nil)=%q", got)
	}
}

// --- helpers ---

func testClient(t *testing.T, baseURL, key string) *Client {
	t.Helper()
	home := t.TempDir()
	writeAuthFile(t, home, `{"apiKey":"`+key+`"}`)
	t.Setenv(APIKeyEnvVar, "")

	client, err := NewClient(home, baseURL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func writeAuthFile(t *testing.T, home, contents string) {
	t.Helper()
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	if err := os.WriteFile(filepath.Join(home, AuthFileName), []byte(contents), 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}
}

// usageFixture is the payload shape observed live for an individual-go account
// mid-billing-period.
func usageFixture(now time.Time) *UsageData {
	return &UsageData{
		Whoami: &Whoami{
			Success: true,
			User:    User{ID: "u1", UserName: "alice", Name: "Alice", Email: "alice@example.test"},
		},
		Credits: &Credits{
			Credits: CreditBalance{MonthlyCredits: 2.408591571},
			WindowLimits: &WindowLimits{
				Limited:  true,
				FiveHour: &WindowSpan{Used: 0, Cap: 3, ResetAt: 0},
				Weekly: &WindowSpan{
					Used:    0.480063398,
					Cap:     6,
					ResetAt: float64(now.Add(7*time.Hour + 4*time.Minute).UnixMilli()),
				},
			},
		},
		Subscription: &Subscription{
			ID:                 "sub_1",
			Status:             "active",
			PlanID:             "individual-go",
			CurrentPeriodStart: "2026-08-14T00:05:12.000Z",
			CurrentPeriodEnd:   "2026-09-14T00:05:12.000Z",
		},
		Summary: &UsageSummary{
			TotalCount: 1040, CompletedCount: 1040, SuccessRate: 100,
			TotalCost: 7.5914084289999995,
		},
		Errors: []string{},
	}
}
