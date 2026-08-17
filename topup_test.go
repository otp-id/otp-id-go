package otpid

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestCreateTopup(t *testing.T) {
	var gotBody map[string]interface{}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/topups" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		raw, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(raw, &gotBody)
		w.Write([]byte(`{"success":true,"data":{"topup_id":"TC20990809Q7M4X2A8BC5D6EFG","payment_url":"https://app.otp.id/topup/TC20990809Q7M4X2A8BC5D6EFG?hash=abc","payment_hash":"abc","amount":100000,"payment_total":100750,"payment_method_id":3,"payment_method":"QRIS","payment_type":"qris","payment_expired_at":"2026-08-14 12:00:00","status":"pending"},"error":null}`))
	})
	res, err := c.CreateTopup(context.Background(), CreateTopupParams{Amount: 100000, PaymentMethodID: 3})
	if err != nil {
		t.Fatal(err)
	}
	if gotBody["amount"] != float64(100000) || gotBody["payment_method_id"] != float64(3) {
		t.Errorf("body = %v", gotBody)
	}
	if res.TopupID != "TC20990809Q7M4X2A8BC5D6EFG" || res.PaymentTotal != 100750 || res.PaymentMethod != "QRIS" {
		t.Errorf("res = %+v", res)
	}
}

func TestCreateTopupInvalidAmount(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success":false,"data":null,"error":{"code":"VALIDATION_ERROR","message":"amount must be one of 10000, 100000, 500000, 1000000, 2000000"}}`))
	})
	_, err := c.CreateTopup(context.Background(), CreateTopupParams{Amount: 12345, PaymentMethodID: 3})
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != ErrCodeValidation {
		t.Fatalf("err = %v, want VALIDATION_ERROR APIError", err)
	}
}
