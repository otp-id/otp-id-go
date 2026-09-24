package otpid

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestOTPStatus(t *testing.T) {
	var gotPath, gotMethod string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Write([]byte(`{"success":true,"data":{"otp_id":"OTP20260807ABCD000001","status":"sent","channel":"whatsapp","number":"6281234567890","attempts":0,"expires_at":"2026-08-07 10:05:00","verified_at":"","price":350},"error":null}`))
	})
	res, err := c.OTPStatus(context.Background(), "OTP20260807ABCD000001")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet || gotPath != "/v3/otp/OTP20260807ABCD000001" {
		t.Errorf("%s %s", gotMethod, gotPath)
	}
	if res.Status != "sent" || res.Attempts != 0 || res.VerifiedAt != "" || res.Price != 350 {
		t.Errorf("res = %+v", res)
	}
	if res.Verification != nil {
		t.Error("non-misscall status must have nil Verification")
	}
	if res.Failure != nil {
		t.Error("successful status must have nil Failure")
	}
}

func TestOTPStatusMisscallPrefix(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"data":{"otp_id":"OTP20260807ABCD000003","status":"sent","channel":"misscall","number":"6281234567890","attempts":1,"expires_at":"2026-08-07 10:05:00","verified_at":"","price":250,"verification":{"prefix":"628559263"}},"error":null}`))
	})
	res, err := c.OTPStatus(context.Background(), "OTP20260807ABCD000003")
	if err != nil {
		t.Fatal(err)
	}
	if res.Verification == nil || res.Verification.Prefix != "628559263" {
		t.Errorf("verification = %+v", res.Verification)
	}
}

func TestOTPStatusInboundVerification(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"data":{"otp_id":"OTP20260807ABCD000002","status":"pending","channel":"whatsapp_inbound","number":"","attempts":0,"expires_at":"2026-08-07 10:05:00","verified_at":"","price":350,"verification":{"wa_number":"6285212345678","message":"OTPID V-8FK2QN9P — verifikasi MyApp. Kirim pesan ini tanpa mengubah isinya.","wa_link":"https://wa.me/6285212345678?text=OTPID%20V-8FK2QN9P","expires_at":"2026-08-07 10:05:00"}},"error":null}`))
	})
	res, err := c.OTPStatus(context.Background(), "OTP20260807ABCD000002")
	if err != nil {
		t.Fatal(err)
	}
	v := res.Verification
	if v == nil {
		t.Fatal("pending whatsapp_inbound status must have a Verification block")
	}
	if v.WaNumber != "6285212345678" || v.WaLink == "" || v.Message == "" || v.ExpiresAt != "2026-08-07 10:05:00" {
		t.Errorf("verification = %+v", v)
	}
	if res.Failure != nil {
		t.Error("pending status must have nil Failure")
	}
}

func TestOTPStatusFailure(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"data":{"otp_id":"OTP20260807ABCD000004","status":"failed","channel":"whatsapp","number":"6281234567890","attempts":0,"expires_at":"2026-08-07 10:05:00","verified_at":"","price":350,"failure":{"code":"PROVIDER_UNAVAILABLE","message":"Penyedia layanan sedang tidak tersedia"}},"error":null}`))
	})
	res, err := c.OTPStatus(context.Background(), "OTP20260807ABCD000004")
	if err != nil {
		t.Fatal(err)
	}
	if res.Failure == nil || res.Failure.Code != FailureCodeProviderUnavailable || res.Failure.Message == "" {
		t.Errorf("Failure = %+v", res.Failure)
	}
}

func TestOTPStatusPathEscapesOtpID(t *testing.T) {
	var gotEscaped string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotEscaped = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success":false,"data":null,"error":{"code":"OTP_NOT_FOUND","message":"not found"}}`))
	})
	_, err := c.OTPStatus(context.Background(), "weird/../id")
	if err == nil {
		t.Fatal("want APIError")
	}
	if !strings.Contains(gotEscaped, "weird%2F..%2Fid") {
		t.Errorf("otp_id must be path-escaped, got %q", gotEscaped)
	}
}

func TestOTPStatusEmptyOtpID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("no network call must happen with an empty otp_id")
	})
	_, err := c.OTPStatus(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "otp_id is empty") {
		t.Fatalf("err = %v, want empty otp_id error", err)
	}
}
