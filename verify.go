package otpid

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var errEmptyOtpID = errors.New("otpid: otp_id is empty")

type verifyOTPBody struct {
	OtpID string `json:"otp_id"`
	Otp   string `json:"otp"`
}

// VerifyResult is the success payload of POST /v3/verify.
type VerifyResult struct {
	OtpID    string `json:"otp_id"`
	Verified bool   `json:"verified"`
	// Reason is "" when Verified is true, "mismatch" when the code was
	// wrong. A mismatch is HTTP 200 and therefore NOT an error from this
	// method. Expired / locked / already-used transactions come back as
	// *APIError (OTP_EXPIRED, TOO_MANY_ATTEMPTS, ALREADY_USED) instead.
	Reason string `json:"reason"`
}

// VerifyOTP checks a user-submitted code against a transaction
// (POST /v3/verify). Do not call it for whatsapp_inbound transactions.
func (c *Client) VerifyOTP(ctx context.Context, otpID, otp string) (*VerifyResult, error) {
	id := strings.TrimSpace(otpID)
	if id == "" {
		return nil, errEmptyOtpID
	}
	var out VerifyResult
	if err := c.doRequest(ctx, http.MethodPost, "/v3/verify", verifyOTPBody{OtpID: id, Otp: otp}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
