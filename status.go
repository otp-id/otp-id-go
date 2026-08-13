package otpid

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// StatusResult is the success payload of GET /v3/otp/{otp_id}.
type StatusResult struct {
	OtpID      string  `json:"otp_id"`
	Status     string  `json:"status"` // sent | success | failed | pending | verified
	Channel    Channel `json:"channel"`
	Number     string  `json:"number"`
	Attempts   int     `json:"attempts"`
	ExpiresAt  string  `json:"expires_at"`
	VerifiedAt string  `json:"verified_at"` // "" until verified
	Price      int     `json:"price"`
	// Verification is only present for not-yet-verified misscall
	// transactions (Prefix field), so polling clients can build their UI.
	Verification *Verification `json:"verification,omitempty"`
}

// OTPStatus fetches the current state of a transaction
// (GET /v3/otp/{otp_id}).
func (c *Client) OTPStatus(ctx context.Context, otpID string) (*StatusResult, error) {
	id := strings.TrimSpace(otpID)
	if id == "" {
		return nil, errEmptyOtpID
	}
	var out StatusResult
	if err := c.doRequest(ctx, http.MethodGet, "/v3/otp/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
