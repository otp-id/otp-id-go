package otpid

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

// Known-good vector, independently precomputed:
//
//	HMAC-SHA256("whsec_testsecret", "1765700000" + "." + webhookBody)
const (
	webhookSecret    = "whsec_testsecret"
	webhookTimestamp = "1765700000"
	webhookSignature = "41c831b6192fa304f0564bd03bb147587119a97a579bb7a0fc12ae9b2c8ed4ca"
)

var webhookBody = []byte(`{"event":"otp.verified","otp_id":"OTP20260807ABCD000001","external_id":"order-8821","channel":"whatsapp","number":"6281234567890","verified_at":"2026-08-07 10:01:30"}`)

// stubNow freezes the package clock near the vector timestamp.
func stubNow(t *testing.T, at time.Time) {
	t.Helper()
	orig := timeNow
	timeNow = func() time.Time { return at }
	t.Cleanup(func() { timeNow = orig })
}

func TestVerifyWebhookSignatureVector(t *testing.T) {
	if !VerifyWebhookSignature(webhookSecret, webhookTimestamp, webhookBody, webhookSignature) {
		t.Fatal("known-good vector must verify")
	}
}

func TestVerifyWebhookSignatureMatchesReferenceHMAC(t *testing.T) {
	// Cross-check against an inline reference implementation so the SDK
	// cannot silently drift from the server's signing scheme.
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write([]byte(webhookTimestamp + "."))
	mac.Write(webhookBody)
	ref := hex.EncodeToString(mac.Sum(nil))
	if !VerifyWebhookSignature(webhookSecret, webhookTimestamp, webhookBody, ref) {
		t.Fatal("reference HMAC must verify")
	}
}

func TestVerifyWebhookSignatureRejects(t *testing.T) {
	cases := map[string]struct {
		secret, ts, sig string
		body            []byte
	}{
		"wrong secret":    {"other-secret", webhookTimestamp, webhookSignature, webhookBody},
		"wrong timestamp": {webhookSecret, "1765700001", webhookSignature, webhookBody},
		"tampered body":   {webhookSecret, webhookTimestamp, webhookSignature, []byte(`{"event":"otp.verified","otp_id":"HACKED"}`)},
		"wrong signature": {webhookSecret, webhookTimestamp, "deadbeef", webhookBody},
	}
	for name, tc := range cases {
		if VerifyWebhookSignature(tc.secret, tc.ts, tc.body, tc.sig) {
			t.Errorf("%s: must not verify", name)
		}
	}
}

func TestParseVerifiedEvent(t *testing.T) {
	stubNow(t, time.Unix(1765700000, 0).Add(30*time.Second))
	ev, err := ParseVerifiedEvent(webhookSecret, webhookTimestamp, webhookSignature, webhookBody)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Event != "otp.verified" || ev.OtpID != "OTP20260807ABCD000001" ||
		ev.ExternalID != "order-8821" || ev.Channel != ChannelWhatsApp ||
		ev.Number != "6281234567890" || ev.VerifiedAt != "2026-08-07 10:01:30" {
		t.Errorf("ev = %+v", ev)
	}
}

func TestParseVerifiedEventInvalidSignature(t *testing.T) {
	stubNow(t, time.Unix(1765700000, 0))
	_, err := ParseVerifiedEvent("other-secret", webhookTimestamp, webhookSignature, webhookBody)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("err = %v, want ErrInvalidSignature", err)
	}
}

func TestParseVerifiedEventStaleTimestamp(t *testing.T) {
	// 6 minutes after the signed timestamp: outside the 5-minute tolerance.
	stubNow(t, time.Unix(1765700000, 0).Add(6*time.Minute))
	_, err := ParseVerifiedEvent(webhookSecret, webhookTimestamp, webhookSignature, webhookBody)
	if !errors.Is(err, ErrStaleTimestamp) {
		t.Fatalf("err = %v, want ErrStaleTimestamp", err)
	}
}

func TestParseVerifiedEventFutureTimestamp(t *testing.T) {
	// Clock skew guard also applies in the other direction.
	stubNow(t, time.Unix(1765700000, 0).Add(-6*time.Minute))
	_, err := ParseVerifiedEvent(webhookSecret, webhookTimestamp, webhookSignature, webhookBody)
	if !errors.Is(err, ErrStaleTimestamp) {
		t.Fatalf("err = %v, want ErrStaleTimestamp", err)
	}
}

func TestParseVerifiedEventNonNumericTimestamp(t *testing.T) {
	stubNow(t, time.Unix(1765700000, 0))
	ts := "not-a-number"
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write([]byte(ts + "."))
	mac.Write(webhookBody)
	sig := hex.EncodeToString(mac.Sum(nil))
	_, err := ParseVerifiedEvent(webhookSecret, ts, sig, webhookBody)
	if !errors.Is(err, ErrStaleTimestamp) {
		t.Fatalf("err = %v, want ErrStaleTimestamp", err)
	}
}

func TestParseVerifiedEventBadJSON(t *testing.T) {
	stubNow(t, time.Unix(1765700000, 0))
	body := []byte(`{not-json`)
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write([]byte(webhookTimestamp + "."))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	_, err := ParseVerifiedEvent(webhookSecret, webhookTimestamp, sig, body)
	if err == nil || errors.Is(err, ErrInvalidSignature) || errors.Is(err, ErrStaleTimestamp) {
		t.Fatalf("err = %v, want a decode error", err)
	}
}
