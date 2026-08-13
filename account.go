package otpid

import (
	"context"
	"net/http"
)

// AccountResult is the success payload of GET /v3/account. It never
// contains credentials.
type AccountResult struct {
	MerchantID string `json:"merchant_id"`
	Name       string `json:"name"`
	BrandName  string `json:"brand_name"`
	BrandEmail string `json:"brand_email"`
	Email      string `json:"email"`
	Saldo      int    `json:"saldo"` // current credit balance
}

// Account fetches the merchant profile and credit balance for the API key
// in use (GET /v3/account).
func (c *Client) Account(ctx context.Context) (*AccountResult, error) {
	var out AccountResult
	if err := c.doRequest(ctx, http.MethodGet, "/v3/account", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
