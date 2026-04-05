package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetFilteredUsageEvents(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/events_response.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dashboard/get-filtered-usage-events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type, got %q", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(fixture)
	}))
	defer srv.Close()

	client := NewClient("test-token", WithBaseURL(srv.URL))
	resp, err := client.GetFilteredUsageEvents(EventsRequest{
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TotalUsageEventsCount != 30653 {
		t.Errorf("expected 30653 total events, got %d", resp.TotalUsageEventsCount)
	}
	if len(resp.UsageEventsDisplay) != 2 {
		t.Fatalf("expected 2 events, got %d", len(resp.UsageEventsDisplay))
	}

	first := resp.UsageEventsDisplay[0]
	if first.Model != "claude-4.6-opus-high-thinking" {
		t.Errorf("expected model claude-4.6-opus-high-thinking, got %q", first.Model)
	}
	if first.ChargedCents != 124.73 {
		t.Errorf("expected chargedCents 124.73, got %f", first.ChargedCents)
	}
	if first.TokenUsage.OutputTokens != 20525 {
		t.Errorf("expected 20525 output tokens, got %d", first.TokenUsage.OutputTokens)
	}
	if first.IsHeadless {
		t.Error("expected first event to not be headless")
	}

	second := resp.UsageEventsDisplay[1]
	if !second.IsHeadless {
		t.Error("expected second event to be headless")
	}
	if second.Kind != "USAGE_EVENT_KIND_INCLUDED_IN_BUSINESS" {
		t.Errorf("unexpected kind: %s", second.Kind)
	}
}

func TestGetFilteredUsageEventsSendsRequestBody(t *testing.T) {
	var receivedBody EventsRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.Write([]byte(`{"totalUsageEventsCount": 0, "usageEventsDisplay": []}`))
	}))
	defer srv.Close()

	client := NewClient("tok", WithBaseURL(srv.URL))
	_, err := client.GetFilteredUsageEvents(EventsRequest{
		TeamID:    2168997,
		UserID:    152683922,
		StartDate: "1774846800000",
		EndDate:   "1775451599999",
		Page:      2,
		PageSize:  50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody.TeamID != 2168997 {
		t.Errorf("expected teamId 2168997, got %d", receivedBody.TeamID)
	}
	if receivedBody.Page != 2 {
		t.Errorf("expected page 2, got %d", receivedBody.Page)
	}
	if receivedBody.StartDate != "1774846800000" {
		t.Errorf("expected startDate 1774846800000, got %q", receivedBody.StartDate)
	}
}

func TestGetFilteredUsageEventsAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "not_authenticated"}`))
	}))
	defer srv.Close()

	client := NewClient("bad", WithBaseURL(srv.URL))
	_, err := client.GetFilteredUsageEvents(EventsRequest{})
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
