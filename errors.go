// Package otpid is the official Go SDK for the OTP.ID V3 API.
package otpid

import "fmt"

// Error codes returned by the OTP.ID V3 API (kept in sync with the server).
const (
	ErrCodeUnauthorized           = "UNAUTHORIZED"
	ErrCodeValidation             = "VALIDATION_ERROR"
	ErrCodeInvalidChannel         = "INVALID_CHANNEL"
	ErrCodeInvalidNumber          = "INVALID_NUMBER"
	ErrCodeInsufficientBalance    = "INSUFFICIENT_BALANCE"
	ErrCodeOtpNotFound            = "OTP_NOT_FOUND"
	ErrCodeDuplicateExternalID    = "DUPLICATE_EXTERNAL_ID"
	ErrCodeOtpExpired             = "OTP_EXPIRED"
	ErrCodeTooManyAttempts        = "TOO_MANY_ATTEMPTS"
	ErrCodeAlreadyUsed            = "ALREADY_USED"
	ErrCodeRateLimited            = "RATE_LIMITED"
	ErrCodeDestinationRateLimited = "DESTINATION_RATE_LIMITED"
	ErrCodeChannelUnavailable     = "CHANNEL_UNAVAILABLE"
	ErrCodeIPNotAllowed           = "IP_NOT_ALLOWED"
	ErrCodeInternal               = "INTERNAL_ERROR"

	// ErrCodeInvalidResponse is produced by the SDK itself (never by the
	// server) when a response body cannot be decoded as a V3 JSON envelope.
	ErrCodeInvalidResponse = "INVALID_RESPONSE"
)

// APIError is the error returned for any non-success API response.
// Match it with errors.As and switch on Code.
type APIError struct {
	Code       string
	Message    string
	Details    map[string]interface{} // e.g. {"existing_otp_id": "..."} on DUPLICATE_EXTERNAL_ID
	HTTPStatus int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("otpid: %s: %s (http %d)", e.Code, e.Message, e.HTTPStatus)
}
