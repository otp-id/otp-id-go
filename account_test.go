package otpid

import (
	"context"
	"net/http"
	"testing"
)

func TestAccount(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/account" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"success":true,"data":{"merchant_id":"M123","name":"PT Contoh","brand_name":"MyApp","brand_email":"otp@myapp.co.id","email":"owner@myapp.co.id","saldo":99650},"error":null}`))
	})
	res, err := c.Account(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.MerchantID != "M123" || res.BrandName != "MyApp" || res.Saldo != 99650 {
		t.Errorf("res = %+v", res)
	}
}

func TestAccountUnauthorized(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success":false,"data":null,"error":{"code":"UNAUTHORIZED","message":"invalid api key"}}`))
	})
	_, err := c.Account(context.Background())
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != ErrCodeUnauthorized || apiErr.HTTPStatus != 401 {
		t.Fatalf("err = %v, want UNAUTHORIZED APIError http 401", err)
	}
}
