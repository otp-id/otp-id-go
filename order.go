package otpid

import (
	"context"
	"net/http"
)

// OrderParams is the request body for RequestOTP and SendOTP.
type OrderParams struct {
	// Channel selects the delivery channel. Required.
	Channel Channel `json:"channel"`
	// Destination is the phone number (digits only) or email address.
	// Required for every channel except ChannelWhatsAppInbound.
	Destination string `json:"destination,omitempty"`
	// Brand overrides the merchant brand_name shown in the OTP message.
	// Required by the server for ChannelVoice.
	Brand string `json:"brand,omitempty"`
	// OtpLength is the generated code length. Server default 6, clamped 4-8.
	// The server forces 4 for ChannelVoice.
	OtpLength int `json:"otp_length,omitempty"`
	// TTL is the OTP validity in seconds. Server default 300, clamped 60-900.
	TTL int `json:"ttl,omitempty"`
	// ExternalID is an optional merchant-side idempotency key.
	ExternalID string `json:"external_id,omitempty"`
}

// Verification is the flat superset of the per-channel `verification`
// response block. Which fields are set depends on the channel:
// whatsapp_inbound fills WaNumber/Message/WaLink/ExpiresAt; misscall fills
// Prefix (and OtpLength on order responses). All other channels have no
// verification block at all (nil pointer).
type Verification struct {
	WaNumber  string `json:"wa_number,omitempty"`
	Message   string `json:"message,omitempty"`
	WaLink    string `json:"wa_link,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	OtpLength int    `json:"otp_length,omitempty"`
}

// OrderResult is the success payload of POST /v3/request and POST /v3/send.
type OrderResult struct {
	OtpID string `json:"otp_id"`
	// Status is one of: sent | success | failed | pending.
	// whatsapp_inbound always starts as "pending" (not charged until
	// verified); "pending" may appear for some vendor deliveries. An
	// idempotency replay reflects the current state, so it may also be
	// "verified" or (whatsapp_inbound only) "expired".
	Status  string  `json:"status"`
	Channel Channel `json:"channel"`
	Number  string  `json:"number"`
	// Price is the credit cost of this transaction. For whatsapp_inbound it
	// is the amount that will be charged only if the transaction is
	// verified; nothing is deducted at request time.
	Price int `json:"price"`
	// LastBalance is the remaining credit after this transaction. It stays
	// unchanged when delivery failed (status "failed"), on an idempotency
	// replay, and for whatsapp_inbound (charged only when verified).
	LastBalance int `json:"last_balance"`
	// ExpiresAt is "YYYY-MM-DD HH:MM:SS" in WIB (UTC+7). Kept as a string;
	// the SDK does not parse server datetimes.
	ExpiresAt    string        `json:"expires_at"`
	Verification *Verification `json:"verification,omitempty"`
	// Failure describes why delivery failed. Set only when Status ==
	// "failed"; nil otherwise. See the FailureCode* constants — treat any
	// unrecognized Code gracefully.
	Failure *Failure `json:"failure,omitempty"`
}

// RequestOTP creates an OTP transaction with a server-generated code
// (POST /v3/request). The code itself is never returned.
func (c *Client) RequestOTP(ctx context.Context, p OrderParams) (*OrderResult, error) {
	var out OrderResult
	if err := c.doRequest(ctx, http.MethodPost, "/v3/request", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// sendOTPBody adds the required otp field on top of OrderParams.
type sendOTPBody struct {
	OrderParams
	Otp string `json:"otp"`
}

// SendOTP delivers a client-generated code (POST /v3/send). The server
// rejects ChannelVoice and ChannelWhatsAppInbound for this endpoint; use
// ChannelWhatsApp, ChannelSMS, or ChannelEmail.
func (c *Client) SendOTP(ctx context.Context, otp string, p OrderParams) (*OrderResult, error) {
	var out OrderResult
	if err := c.doRequest(ctx, http.MethodPost, "/v3/send", sendOTPBody{OrderParams: p, Otp: otp}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
