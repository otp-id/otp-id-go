package otpid

import (
	"context"
	"net/http"
)

// CreateTopupParams is the request body for POST /v3/topups.
type CreateTopupParams struct {
	// Amount is the credit package in rupiah. The server accepts exactly:
	// 10000, 100000, 500000, 1000000, 2000000.
	Amount int `json:"amount"`
	// PaymentMethodID selects the payment method for the invoice.
	PaymentMethodID int `json:"payment_method_id"`
}

// TopupResult is the success payload of POST /v3/topups. PaymentURL is a
// signed OTP.ID payment page that opens without a dashboard login.
type TopupResult struct {
	TopupID          string `json:"topup_id"`
	PaymentURL       string `json:"payment_url"`
	PaymentHash      string `json:"payment_hash"`
	Amount           int    `json:"amount"`
	PaymentTotal     int    `json:"payment_total"` // amount + admin fee; display as-is
	PaymentMethodID  int    `json:"payment_method_id"`
	PaymentMethod    string `json:"payment_method"`
	PaymentType      string `json:"payment_type"`
	PaymentExpiredAt string `json:"payment_expired_at"`
	Status           string `json:"status"`
}

// CreateTopup creates a credit top-up invoice (POST /v3/topups). Call it
// from server-side code only — never expose your API key to browsers or
// mobile apps.
func (c *Client) CreateTopup(ctx context.Context, p CreateTopupParams) (*TopupResult, error) {
	var out TopupResult
	if err := c.doRequest(ctx, http.MethodPost, "/v3/topups", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
