package otpid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient starts an httptest server and returns a Client pointed at it.
// Reused by every endpoint test in this package.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient("test-key", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("  test-key  ")
	if c.apiKey != "test-key" {
		t.Errorf("apiKey = %q, want trimmed %q", c.apiKey, "test-key")
	}
	if c.baseURL != "https://api.otp.id" {
		t.Errorf("baseURL = %q, want https://api.otp.id", c.baseURL)
	}
	if c.httpClient == nil || c.httpClient.Timeout != 30*time.Second {
		t.Error("httpClient default must have 30s timeout")
	}
}

func TestWithBaseURLTrimsTrailingSlash(t *testing.T) {
	c := NewClient("k", WithBaseURL("https://example.com/"))
	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, want without trailing slash", c.baseURL)
	}
}

func TestDoRequestSendsHeaders(t *testing.T) {
	var gotAuth, gotUA, gotCT string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		gotCT = r.Header.Get("Content-Type")
		w.Write([]byte(`{"success":true,"data":{},"error":null}`))
	})
	var out struct{}
	if err := c.doRequest(context.Background(), http.MethodPost, "/v3/request", map[string]string{"a": "b"}, &out); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotUA != "otp-id-go/"+Version {
		t.Errorf("User-Agent = %q", gotUA)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q", gotCT)
	}
}

func TestDoRequestGetHasNoContentType(t *testing.T) {
	var gotCT string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Write([]byte(`{"success":true,"data":{},"error":null}`))
	})
	var out struct{}
	if err := c.doRequest(context.Background(), http.MethodGet, "/v3/account", nil, &out); err != nil {
		t.Fatal(err)
	}
	if gotCT != "" {
		t.Errorf("GET must not send Content-Type, got %q", gotCT)
	}
}

func TestDoRequestAPIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"success":false,"data":null,"error":{"code":"DUPLICATE_EXTERNAL_ID","message":"external_id already used","details":{"existing_otp_id":"OTP20260807ABCD000001"}}}`))
	})
	err := c.doRequest(context.Background(), http.MethodPost, "/v3/request", map[string]string{}, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T, want *APIError", err)
	}
	if apiErr.Code != ErrCodeDuplicateExternalID || apiErr.HTTPStatus != 409 {
		t.Errorf("got %+v", apiErr)
	}
	if apiErr.Details["existing_otp_id"] != "OTP20260807ABCD000001" {
		t.Errorf("Details = %v", apiErr.Details)
	}
}

func TestDoRequestNonJSONResponse(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("<html>502 Bad Gateway</html>"))
	})
	err := c.doRequest(context.Background(), http.MethodGet, "/v3/account", nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T, want *APIError", err)
	}
	if apiErr.Code != ErrCodeInvalidResponse || apiErr.HTTPStatus != 502 {
		t.Errorf("got %+v", apiErr)
	}
	if !strings.Contains(apiErr.Message, "502 Bad Gateway") {
		t.Errorf("Message should carry a body snippet, got %q", apiErr.Message)
	}
}

func TestDoRequestFailWithoutErrorBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"data":null,"error":null}`))
	})
	err := c.doRequest(context.Background(), http.MethodGet, "/v3/account", nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T, want *APIError", err)
	}
	if apiErr.Code != ErrCodeInvalidResponse {
		t.Errorf("Code = %q, want INVALID_RESPONSE", apiErr.Code)
	}
}

func TestDoRequestSuccessMissingData(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"data":null,"error":null}`))
	})
	var out struct{}
	err := c.doRequest(context.Background(), http.MethodGet, "/v3/account", nil, &out)
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != ErrCodeInvalidResponse {
		t.Fatalf("want INVALID_RESPONSE for null data with out, got %v", err)
	}
}

func TestDoRequestEmptyAPIKey(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	t.Cleanup(srv.Close)
	c := NewClient("   ", WithBaseURL(srv.URL))
	err := c.doRequest(context.Background(), http.MethodGet, "/v3/account", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "api key is empty") {
		t.Fatalf("err = %v, want empty api key error", err)
	}
	if called {
		t.Error("no network call must happen with an empty api key")
	}
}

func TestDoRequestContextCanceled(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"data":{},"error":null}`))
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.doRequest(ctx, http.MethodGet, "/v3/account", nil, nil)
	if err == nil {
		t.Fatal("want error from canceled context")
	}
}
