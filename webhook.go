package otpid

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// WebhookTolerance is the maximum accepted clock difference between the
// X-OTPID-Timestamp header and the local clock in ParseVerifiedEvent.
const WebhookTolerance = 5 * time.Minute

var (
	// ErrInvalidSignature is returned when the X-OTPID-Signature header
	// does not match the payload.
	ErrInvalidSignature = errors.New("otpid: invalid webhook signature")
	// ErrStaleTimestamp is returned when X-OTPID-Timestamp is not a unix
	// timestamp within WebhookTolerance of the local clock.
	ErrStaleTimestamp = errors.New("otpid: webhook timestamp outside tolerance")
)

// timeNow is stubbed in tests.
var timeNow = time.Now

// VerifiedEvent is the payload of the otp.verified webhook.
type VerifiedEvent struct {
	Event      string  `json:"event"` // always "otp.verified"
	OtpID      string  `json:"otp_id"`
	ExternalID string  `json:"external_id"` // "" when the merchant sent no external_id
	Channel    Channel `json:"channel"`
	Number     string  `json:"number"`
	VerifiedAt string  `json:"verified_at"` // "YYYY-MM-DD HH:MM:SS" WIB
}

func computeWebhookSignature(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyWebhookSignature reports whether signature matches
// hex(HMAC-SHA256(secret, timestamp + "." + body)). The comparison is
// constant-time. It performs no timestamp freshness check — use
// ParseVerifiedEvent for the full validation.
func VerifyWebhookSignature(secret, timestamp string, body []byte, signature string) bool {
	expected := computeWebhookSignature(secret, timestamp, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ParseVerifiedEvent validates an incoming otp.verified webhook and
// returns its payload. It checks the signature (constant-time), then the
// timestamp freshness (±WebhookTolerance, anti-replay), then decodes the
// body. Pass the raw request body and the X-OTPID-Timestamp /
// X-OTPID-Signature header values unmodified.
func ParseVerifiedEvent(secret, timestamp, signature string, body []byte) (*VerifiedEvent, error) {
	if !VerifyWebhookSignature(secret, timestamp, body, signature) {
		return nil, ErrInvalidSignature
	}
	ts, err := strconv.ParseInt(strings.TrimSpace(timestamp), 10, 64)
	if err != nil {
		return nil, ErrStaleTimestamp
	}
	diff := timeNow().Sub(time.Unix(ts, 0))
	if diff > WebhookTolerance || diff < -WebhookTolerance {
		return nil, ErrStaleTimestamp
	}
	var ev VerifiedEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		return nil, fmt.Errorf("otpid: decode webhook payload: %w", err)
	}
	return &ev, nil
}
