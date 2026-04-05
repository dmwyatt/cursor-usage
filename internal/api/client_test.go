package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAttachesCookie(t *testing.T) {
	var gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("WorkosCursorSessionToken")
		if err == nil {
			gotCookie = c.Value
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	client := NewClient("my-secret-token", WithBaseURL(srv.URL))
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/usage-summary", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if gotCookie != "my-secret-token" {
		t.Errorf("expected cookie %q, got %q", "my-secret-token", gotCookie)
	}
}

func TestClientSetsOriginOnPOST(t *testing.T) {
	var gotOrigin string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotOrigin = r.Header.Get("Origin")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	client := NewClient("tok", WithBaseURL(srv.URL))
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/dashboard/get-filtered-usage-events", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if gotOrigin != srv.URL {
		t.Errorf("expected Origin %q, got %q", srv.URL, gotOrigin)
	}
}

func TestClientDoesNotSetOriginOnGET(t *testing.T) {
	var gotOrigin string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotOrigin = r.Header.Get("Origin")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	client := NewClient("tok", WithBaseURL(srv.URL))
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/usage-summary", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if gotOrigin != "" {
		t.Errorf("expected no Origin header on GET, got %q", gotOrigin)
	}
}

func TestClient401ReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "not_authenticated"}`))
	}))
	defer srv.Close()

	client := NewClient("bad-token", WithBaseURL(srv.URL))
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/usage-summary", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error from Do: %v", err)
	}

	_, err = client.CheckResponse(resp)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestClient429ReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`rate limited`))
	}))
	defer srv.Close()

	client := NewClient("tok", WithBaseURL(srv.URL))
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error from Do: %v", err)
	}

	_, err = client.CheckResponse(resp)
	if err == nil {
		t.Fatal("expected error for 429, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 429 {
		t.Errorf("expected status 429, got %d", apiErr.StatusCode)
	}
}

func TestClientSuccessfulResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key": "value"}`))
	}))
	defer srv.Close()

	client := NewClient("tok", WithBaseURL(srv.URL))
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error from Do: %v", err)
	}

	body, err := client.CheckResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := io.ReadAll(body)
	if string(data) != `{"key": "value"}` {
		t.Errorf("unexpected body: %s", data)
	}
}
