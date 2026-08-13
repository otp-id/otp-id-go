package otpid

import (
	"errors"
	"testing"
)

func TestAPIErrorError(t *testing.T) {
	err := &APIError{Code: ErrCodeInsufficientBalance, Message: "saldo tidak cukup", HTTPStatus: 402}
	want := "otpid: INSUFFICIENT_BALANCE: saldo tidak cukup (http 402)"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIErrorAs(t *testing.T) {
	var wrapped error = &APIError{Code: ErrCodeOtpExpired, HTTPStatus: 422}
	var apiErr *APIError
	if !errors.As(wrapped, &apiErr) {
		t.Fatal("errors.As should match *APIError")
	}
	if apiErr.Code != "OTP_EXPIRED" {
		t.Errorf("Code = %q, want OTP_EXPIRED", apiErr.Code)
	}
}

func TestErrorCodeValues(t *testing.T) {
	// Guards the constants against typos — values are the API contract.
	cases := map[string]string{
		ErrCodeUnauthorized:           "UNAUTHORIZED",
		ErrCodeValidation:             "VALIDATION_ERROR",
		ErrCodeInvalidChannel:         "INVALID_CHANNEL",
		ErrCodeInvalidNumber:          "INVALID_NUMBER",
		ErrCodeInsufficientBalance:    "INSUFFICIENT_BALANCE",
		ErrCodeOtpNotFound:            "OTP_NOT_FOUND",
		ErrCodeDuplicateExternalID:    "DUPLICATE_EXTERNAL_ID",
		ErrCodeOtpExpired:             "OTP_EXPIRED",
		ErrCodeTooManyAttempts:        "TOO_MANY_ATTEMPTS",
		ErrCodeAlreadyUsed:            "ALREADY_USED",
		ErrCodeRateLimited:            "RATE_LIMITED",
		ErrCodeDestinationRateLimited: "DESTINATION_RATE_LIMITED",
		ErrCodeChannelUnavailable:     "CHANNEL_UNAVAILABLE",
		ErrCodeIPNotAllowed:           "IP_NOT_ALLOWED",
		ErrCodeInternal:               "INTERNAL_ERROR",
		ErrCodeInvalidResponse:        "INVALID_RESPONSE",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("constant = %q, want %q", got, want)
		}
	}
}
