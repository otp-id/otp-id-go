package otpid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Version is the SDK version, sent in the User-Agent header.
const Version = "0.1.1"

const (
	defaultBaseURL = "https://api.otp.id"
	maxBodyBytes   = 1 << 20 // 1 MB guard against abnormal responses
)

var errEmptyAPIKey = errors.New("otpid: api key is empty")

// Client is an OTP.ID V3 API client. It is safe for concurrent use.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default base URL (https://api.otp.id).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient replaces the default *http.Client (30s timeout).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// NewClient creates a Client authenticated with the given API key.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// envelope is the V3 response wrapper: {success, data, error}.
type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *errorBody      `json:"error"`
}

type errorBody struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

// doRequest performs a single HTTP call (no retries), decodes the V3
// envelope, and unmarshals data into out when out is non-nil.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	if c.apiKey == "" {
		return errEmptyAPIKey
	}
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("otpid: encode request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("otpid: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "otp-id-go/"+Version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("otpid: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return fmt.Errorf("otpid: read response body: %w", err)
	}
	invalid := func() *APIError {
		return &APIError{Code: ErrCodeInvalidResponse, Message: bodySnippet(raw), HTTPStatus: resp.StatusCode}
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return invalid()
	}
	if !env.Success {
		if env.Error == nil {
			return invalid()
		}
		return &APIError{
			Code:       env.Error.Code,
			Message:    env.Error.Message,
			Details:    env.Error.Details,
			HTTPStatus: resp.StatusCode,
		}
	}
	if out != nil {
		if len(env.Data) == 0 || string(env.Data) == "null" {
			return invalid()
		}
		if err := json.Unmarshal(env.Data, out); err != nil {
			return invalid()
		}
	}
	return nil
}

// bodySnippet returns the first ~200 characters of a raw body for
// diagnostics, without splitting multi-byte runes.
func bodySnippet(raw []byte) string {
	const max = 200
	s := strings.TrimSpace(string(raw))
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}
