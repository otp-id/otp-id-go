package otpid

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestVerifyOTPSuccess(t *testing.T) {
	var gotBody map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/verify" {
			t.Errorf("path = %q", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &gotBody)
		w.Write([]byte(`{"success":true,"data":{"otp_id":"OTP20260807ABCD000001","verified":true,"reason":""},"error":null}`))
	})
	res, err := c.VerifyOTP(context.Background(), "OTP20260807ABCD000001", "482913")
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["otp_id"] != "OTP20260807ABCD000001" || gotBody["otp"] != "482913" {
		t.Errorf("body = %v", gotBody)
	}
	if !res.Verified || res.Reason != "" {
		t.Errorf("res = %+v", res)
	}
}

func TestVerifyOTPMismatchIsNotAnError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"data":{"otp_id":"OTP20260807ABCD000001","verified":false,"reason":"mismatch"},"error":null}`))
	})
	res, err := c.VerifyOTP(context.Background(), "OTP20260807ABCD000001", "000000")
	if err != nil {
		t.Fatalf("mismatch must not be an error, got %v", err)
	}
	if res.Verified || res.Reason != "mismatch" {
		t.Errorf("res = %+v", res)
	}
}

func TestVerifyOTPExpiredIsAPIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"success":false,"data":null,"error":{"code":"OTP_EXPIRED","message":"otp has expired"}}`))
	})
	_, err := c.VerifyOTP(context.Background(), "OTP20260807ABCD000001", "482913")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != ErrCodeOtpExpired || apiErr.HTTPStatus != 422 {
		t.Fatalf("err = %v, want OTP_EXPIRED APIError http 422", err)
	}
}

func TestVerifyOTPEmptyOtpID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("no network call must happen with an empty otp_id")
	})
	_, err := c.VerifyOTP(context.Background(), "  ", "482913")
	if err == nil || !strings.Contains(err.Error(), "otp_id is empty") {
		t.Fatalf("err = %v, want empty otp_id error", err)
	}
}
