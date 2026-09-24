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
	// Verification is present for not-yet-verified transactions that carry
	// interactive delivery state, so polling clients can build their UI:
	// misscall (Prefix field), and pending, not-expired whatsapp_inbound
	// (WaNumber/Message/WaLink/ExpiresAt fields). Nil for every other
	// channel or once the transaction leaves that state.
	Verification *Verification `json:"verification,omitempty"`
	// Failure describes why delivery failed. Set only when Status ==
	// "failed"; nil otherwise. See the FailureCode* constants — treat any
	// unrecognized Code gracefully.
	Failure *Failure `json:"failure,omitempty"`
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
