package api

import "fmt"

// UsageSummary is the response from GET /api/usage-summary.
type UsageSummary struct {
	BillingCycleStart                  string          `json:"billingCycleStart"`
	BillingCycleEnd                    string          `json:"billingCycleEnd"`
	MembershipType                     string          `json:"membershipType"`
	LimitType                          string          `json:"limitType"`
	IsUnlimited                        bool            `json:"isUnlimited"`
	AutoModelSelectedDisplayMessage    string          `json:"autoModelSelectedDisplayMessage"`
	NamedModelSelectedDisplayMessage   string          `json:"namedModelSelectedDisplayMessage"`
	IndividualUsage                    IndividualUsage `json:"individualUsage"`
	TeamUsage                          TeamUsage       `json:"teamUsage"`
}

// IndividualUsage contains plan and on-demand usage for the authenticated user.
type IndividualUsage struct {
	Plan     PlanUsage     `json:"plan"`
	OnDemand OnDemandUsage `json:"onDemand"`
}

// PlanUsage tracks request-based usage against plan allowance.
type PlanUsage struct {
	Enabled         bool          `json:"enabled"`
	Used            int           `json:"used"`
	Limit           int           `json:"limit"`
	Remaining       int           `json:"remaining"`
	Breakdown       PlanBreakdown `json:"breakdown"`
	AutoPercentUsed int           `json:"autoPercentUsed"`
	APIPercentUsed  int           `json:"apiPercentUsed"`
	TotalPercentUsed int          `json:"totalPercentUsed"`
}

// PlanBreakdown shows included, bonus, and total request allowances.
type PlanBreakdown struct {
	Included int `json:"included"`
	Bonus    int `json:"bonus"`
	Total    int `json:"total"`
}

// OnDemandUsage tracks usage-based (pay-per-use) consumption.
// Limit and Remaining are nullable (nil means unlimited).
type OnDemandUsage struct {
	Enabled   bool `json:"enabled"`
	Used      int  `json:"used"`
	Limit     *int `json:"limit"`
	Remaining *int `json:"remaining"`
}

// TeamUsage contains team-wide spend limits and usage.
type TeamUsage struct {
	OnDemand TeamOnDemandUsage `json:"onDemand"`
}

// TeamOnDemandUsage tracks team-level on-demand spending (values in cents).
type TeamOnDemandUsage struct {
	Enabled   bool    `json:"enabled"`
	Used      float64 `json:"used"`
	Limit     float64 `json:"limit"`
	Remaining float64 `json:"remaining"`
}

// EventsRequest is the POST body for /api/dashboard/get-filtered-usage-events.
type EventsRequest struct {
	TeamID    int    `json:"teamId,omitempty"`
	UserID    int    `json:"userId,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
}

// EventsResponse is the response from the filtered usage events endpoint.
type EventsResponse struct {
	TotalUsageEventsCount int          `json:"totalUsageEventsCount"`
	UsageEventsDisplay    []UsageEvent `json:"usageEventsDisplay"`
}

// UsageEvent is a single usage event with cost and token breakdowns.
type UsageEvent struct {
	Timestamp      string     `json:"timestamp"`
	Model          string     `json:"model"`
	Kind           string     `json:"kind"`
	RequestsCosts  float64    `json:"requestsCosts"`
	UsageBasedCosts string    `json:"usageBasedCosts"`
	IsTokenBasedCall bool     `json:"isTokenBasedCall"`
	TokenUsage     TokenUsage `json:"tokenUsage"`
	OwningUser     string     `json:"owningUser"`
	OwningTeam     string     `json:"owningTeam"`
	CursorTokenFee float64    `json:"cursorTokenFee"`
	IsChargeable   bool       `json:"isChargeable"`
	IsHeadless     bool       `json:"isHeadless"`
	ChargedCents   float64    `json:"chargedCents"`
}

// TokenUsage contains the token breakdown for a single usage event.
type TokenUsage struct {
	InputTokens     int     `json:"inputTokens"`
	OutputTokens    int     `json:"outputTokens"`
	CacheWriteTokens int    `json:"cacheWriteTokens"`
	TotalCents      float64 `json:"totalCents"`
}

// APIError represents an error response from the Cursor API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Body)
}
