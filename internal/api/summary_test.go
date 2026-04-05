package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetUsageSummary(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/summary_response.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/usage-summary" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(fixture)
	}))
	defer srv.Close()

	client := NewClient("test-token", WithBaseURL(srv.URL))
	summary, err := client.GetUsageSummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.MembershipType != "enterprise" {
		t.Errorf("expected enterprise, got %q", summary.MembershipType)
	}
	if summary.BillingCycleStart != "2026-04-02T14:11:55.000Z" {
		t.Errorf("unexpected billing start: %s", summary.BillingCycleStart)
	}
	if summary.IndividualUsage.Plan.Used != 2000 {
		t.Errorf("expected plan used 2000, got %d", summary.IndividualUsage.Plan.Used)
	}
	if summary.IndividualUsage.Plan.Limit != 2000 {
		t.Errorf("expected plan limit 2000, got %d", summary.IndividualUsage.Plan.Limit)
	}
	if summary.IndividualUsage.OnDemand.Used != 2309 {
		t.Errorf("expected on-demand used 2309, got %d", summary.IndividualUsage.OnDemand.Used)
	}
	if summary.IndividualUsage.OnDemand.Limit != nil {
		t.Errorf("expected on-demand limit nil, got %v", summary.IndividualUsage.OnDemand.Limit)
	}
	if summary.IndividualUsage.Plan.Breakdown.Bonus != 6121 {
		t.Errorf("expected bonus 6121, got %d", summary.IndividualUsage.Plan.Breakdown.Bonus)
	}
}

func TestGetUsageSummaryAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "not_authenticated"}`))
	}))
	defer srv.Close()

	client := NewClient("bad-token", WithBaseURL(srv.URL))
	_, err := client.GetUsageSummary()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected 401, got %d", apiErr.StatusCode)
	}
}
