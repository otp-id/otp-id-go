package otpid

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

// Fixtures mirror the examples in otp-be/docs/openapi/v3.yaml and the
// public Mintlify docs (otp-id-docs/api-reference).
const orderWhatsAppFixture = `{"success":true,"data":{"otp_id":"OTP20260807ABCD000001","status":"sent","channel":"whatsapp","number":"6281234567890","price":350,"last_balance":99650,"expires_at":"2026-08-07 10:05:00"},"error":null}`

const orderInboundFixture = `{"success":true,"data":{"otp_id":"OTP20260807ABCD000002","status":"pending","channel":"whatsapp_inbound","number":"","price":350,"last_balance":99300,"expires_at":"2026-08-07 10:05:00","verification":{"wa_number":"6285212345678","message":"OTPID V-8FK2QN9P — verifikasi MyApp. Kirim pesan ini tanpa mengubah isinya.","wa_link":"https://wa.me/6285212345678?text=OTPID%20V-8FK2QN9P","expires_at":"2026-08-07 10:05:00"}},"error":null}`

const orderMisscallFixture = `{"success":true,"data":{"otp_id":"OTP20260807ABCD000003","status":"sent","channel":"misscall","number":"6281234567890","price":250,"last_balance":99050,"expires_at":"2026-08-07 10:05:00","verification":{"prefix":"628559263","otp_length":4}},"error":null}`

const orderFailedFixture = `{"success":true,"data":{"otp_id":"OTP20260807ABCD000004","status":"failed","channel":"whatsapp","number":"6281234567890","price":350,"last_balance":99650,"expires_at":"2026-08-07 10:05:00","failure":{"code":"NUMBER_NOT_ON_WHATSAPP","message":"Nomor tidak terdaftar di WhatsApp"}},"error":null}`

func TestRequestOTPWhatsApp(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(raw, &gotBody)
		w.Write([]byte(orderWhatsAppFixture))
	})
	res, err := c.RequestOTP(context.Background(), OrderParams{
		Channel:     ChannelWhatsApp,
		Destination: "6281234567890",
		Brand:       "MyApp",
		OtpLength:   6,
		TTL:         300,
		ExternalID:  "order-8821",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v3/request" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["channel"] != "whatsapp" || gotBody["destination"] != "6281234567890" ||
		gotBody["brand"] != "MyApp" || gotBody["external_id"] != "order-8821" {
		t.Errorf("body = %v", gotBody)
	}
	if gotBody["otp_length"] != float64(6) || gotBody["ttl"] != float64(300) {
		t.Errorf("body numeric fields = %v", gotBody)
	}
	if _, present := gotBody["otp"]; present {
		t.Error("request body must not contain an otp field")
	}
	if res.OtpID != "OTP20260807ABCD000001" || res.Status != "sent" ||
		res.Channel != ChannelWhatsApp || res.Price != 350 || res.LastBalance != 99650 {
		t.Errorf("res = %+v", res)
	}
	if res.Verification != nil {
		t.Error("whatsapp order must have nil Verification")
	}
	if res.Failure != nil {
		t.Error("successful order must have nil Failure")
	}
}

func TestRequestOTPFailure(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(orderFailedFixture))
	})
	res, err := c.RequestOTP(context.Background(), OrderParams{Channel: ChannelWhatsApp, Destination: "6281234567890"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "failed" {
		t.Errorf("Status = %q, want failed", res.Status)
	}
	if res.Failure == nil {
		t.Fatal("failed order must have a non-nil Failure")
	}
	if res.Failure.Code != FailureCodeNumberNotOnWhatsApp || res.Failure.Message == "" {
		t.Errorf("Failure = %+v", res.Failure)
	}
}

func TestRequestOTPOmitsEmptyOptionalFields(t *testing.T) {
	var gotBody map[string]interface{}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(raw, &gotBody)
		w.Write([]byte(orderInboundFixture))
	})
	if _, err := c.RequestOTP(context.Background(), OrderParams{Channel: ChannelWhatsAppInbound}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"destination", "brand", "otp_length", "ttl", "external_id"} {
		if _, present := gotBody[key]; present {
			t.Errorf("empty optional field %q must be omitted", key)
		}
	}
}

func TestRequestOTPInboundVerification(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(orderInboundFixture))
	})
	res, err := c.RequestOTP(context.Background(), OrderParams{Channel: ChannelWhatsAppInbound})
	if err != nil {
		t.Fatal(err)
	}
	v := res.Verification
	if v == nil {
		t.Fatal("Verification must be set for whatsapp_inbound")
	}
	if v.WaNumber != "6285212345678" || v.WaLink == "" || v.Message == "" || v.ExpiresAt != "2026-08-07 10:05:00" {
		t.Errorf("verification = %+v", v)
	}
}

func TestRequestOTPMisscallVerification(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(orderMisscallFixture))
	})
	res, err := c.RequestOTP(context.Background(), OrderParams{Channel: ChannelMisscall, Destination: "6281234567890"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Verification == nil || res.Verification.Prefix != "628559263" || res.Verification.OtpLength != 4 {
		t.Errorf("verification = %+v", res.Verification)
	}
}

func TestSendOTPBodyIncludesOtp(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(raw, &gotBody)
		w.Write([]byte(orderWhatsAppFixture))
	})
	res, err := c.SendOTP(context.Background(), "482913", OrderParams{
		Channel:     ChannelWhatsApp,
		Destination: "6281234567890",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v3/send" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["otp"] != "482913" || gotBody["channel"] != "whatsapp" || gotBody["destination"] != "6281234567890" {
		t.Errorf("body = %v", gotBody)
	}
	if res.OtpID != "OTP20260807ABCD000001" || res.Status != "sent" || res.LastBalance != 99650 {
		t.Errorf("res = %+v", res)
	}
}

func TestRequestOTPInsufficientBalance(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write([]byte(`{"success":false,"data":null,"error":{"code":"INSUFFICIENT_BALANCE","message":"balance is not enough"}}`))
	})
	_, err := c.RequestOTP(context.Background(), OrderParams{Channel: ChannelSMS, Destination: "6281234567890"})
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != ErrCodeInsufficientBalance {
		t.Fatalf("err = %v, want INSUFFICIENT_BALANCE APIError", err)
	}
}
